"""Exercise the publisher's actual inline validator with hostile build artifacts."""
import hashlib
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import textwrap
import unittest

WORKFLOW = Path(__file__).resolve().parents[1] / '.github/workflows/release.yml'


class PublisherValidationTests(unittest.TestCase):
    def setUp(self):
        workflow = WORKFLOW.read_text()
        inline = workflow.split('        shell: python\n        run: |\n', 1)[1]
        self.validator = textwrap.dedent(inline.split('      - name:', 1)[0])
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.root = Path(self.directory.name) / 'release-assets'
        self.root.mkdir()
        self.prefix = 'terraform-provider-packager_0.2.1'
        self.files = [self.prefix + '_' + platform + '_' + arch + '.zip'
                      for platform in ('darwin', 'linux', 'windows') for arch in ('amd64', 'arm64')]
        self.files.append(self.prefix + '_manifest.json')
        for name in self.files:
            (self.root / name).write_bytes(b'fixture')
        self.sums = self.root / (self.prefix + '_SHA256SUMS')
        self.lines = [hashlib.sha256(b'fixture').hexdigest() + '  ' + name for name in self.files]
        self.sums.write_text('\n'.join(self.lines) + '\n')

    def validate(self, tag='v0.2.1'):
        return subprocess.run([sys.executable, '-I', '-c', self.validator],
                              cwd=self.directory.name, env={**os.environ, 'TAG': tag},
                              capture_output=True).returncode

    def test_valid_assets(self):
        self.assertEqual(self.validate(), 0)

    def test_tampered_content(self):
        (self.root / self.files[0]).write_bytes(b'tampered')
        self.assertNotEqual(self.validate(), 0)

    def test_unexpected_executable(self):
        (self.root / 'hook.sh').write_text('echo dangerous')
        self.assertNotEqual(self.validate(), 0)

    def test_traversal_and_duplicate_entries(self):
        self.sums.write_text('\n'.join(self.lines + [self.lines[0]]) + '\n')
        self.assertNotEqual(self.validate(), 0)
        self.sums.write_text(self.lines[0].split('  ')[0] + '  ../outside\n')
        self.assertNotEqual(self.validate(), 0)

    def test_symlink(self):
        path = self.root / self.files[0]
        path.unlink()
        path.symlink_to(self.files[1])
        self.assertNotEqual(self.validate(), 0)

    def test_version_injection(self):
        self.assertNotEqual(self.validate('v0.2.1\nanything'), 0)
        self.assertNotEqual(self.validate('$(id)'), 0)

    def test_missing_asset(self):
        (self.root / self.files[0]).unlink()
        self.assertNotEqual(self.validate(), 0)


if __name__ == '__main__':
    unittest.main()
