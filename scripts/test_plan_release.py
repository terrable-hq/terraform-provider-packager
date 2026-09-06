import os
from pathlib import Path
import subprocess
import tempfile
import unittest

from plan_release import plan


class ReleasePlanTests(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        previous = os.getcwd()
        self.addCleanup(os.chdir, previous)
        os.chdir(self.directory.name)
        self.git('init', '-q')
        self.git('config', 'commit.gpgsign', 'false')
        self.git('config', 'tag.gpgsign', 'false')
        self.git('config', 'user.name', 'Test')
        self.git('config', 'user.email', 'test@example.invalid')
        self.commit('main.go', 'initial')
        self.git('tag', 'v0.2.0')

    def git(self, *args):
        return subprocess.check_output(['git', *args], text=True).strip()

    def commit(self, name, content):
        path = Path(name)
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content)
        self.git('add', '--', name)
        self.git('commit', '-qm', 'change')

    def test_runtime_patch_and_explicit_minor(self):
        self.commit('main.go', 'changed')
        self.assertEqual(plan(), 'v0.2.1')
        self.assertEqual(plan('v0.3.0'), 'v0.3.0')

    def test_docs_skip_and_explicit_force(self):
        self.commit('docs/index.md', 'docs')
        self.assertEqual(plan(), '')
        self.assertEqual(plan('v0.2.1'), 'v0.2.1')

    def test_only_documentation_is_ignored(self):
        for name in ['docs/build.sh', '.github/workflows/release.yml', 'go.mod']:
            with self.subTest(name=name):
                self.commit(name, 'change')
                self.assertEqual(plan(), 'v0.2.1')

    def test_retry_reuses_tag_at_exact_commit(self):
        self.commit('main.go', 'changed')
        self.git('tag', 'v0.2.1')
        self.assertEqual(plan(), 'v0.2.1')
        self.assertEqual(plan('v0.2.1'), 'v0.2.1')

    def test_semver_order_and_ignore_untrusted_tag_formats(self):
        self.git('tag', 'v0.9.0')
        self.git('tag', 'v0.10.0')
        self.git('tag', 'v999.0.0-$(id)')
        self.commit('main.go', 'changed')
        self.assertEqual(plan(), 'v0.10.1')
        with self.assertRaises(ValueError):
            plan('$(id)')

    def test_compares_since_release_not_only_latest_merge(self):
        self.commit('main.go', 'runtime change')
        self.commit('README.md', 'documentation afterwards')
        self.assertEqual(plan(), 'v0.2.1')


if __name__ == '__main__':
    unittest.main()
