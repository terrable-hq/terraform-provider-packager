"""Choose a preview version from protected tags and changes at the checked-out SHA."""
import os
import re
import subprocess

VERSION = re.compile(r"v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\Z")
DOCS = {"README.md", "CHANGELOG.md", "RELEASING.md", "SECURITY.md"}


def git(*args):
    return subprocess.check_output(["git", *args], text=True).strip()


def plan(version=""):
    if version and not VERSION.fullmatch(version):
        raise ValueError("version must be vMAJOR.MINOR.PATCH")
    tags = [tag for tag in git("tag", "--list").splitlines() if VERSION.fullmatch(tag)]
    latest = max(tags, key=lambda tag: tuple(map(int, VERSION.fullmatch(tag).groups())), default=None)
    if version:
        if version in tags and git("rev-list", "-n", "1", version) == git("rev-parse", "HEAD"):
            return version
        if latest and tuple(map(int, VERSION.fullmatch(version).groups())) <= tuple(map(int, VERSION.fullmatch(latest).groups())):
            raise ValueError("version must exceed existing tags")
        return version
    if latest:
        # A failed publisher may already have reserved this immutable tag.
        if git("rev-list", "-n", "1", latest) == git("rev-parse", "HEAD"):
            return latest
        names = git("diff", "--name-only", latest, "HEAD", "--").splitlines()
        docs_only = all(name in DOCS or (name.startswith("docs/") and name.endswith(".md")) for name in names)
        if docs_only:
            return ""
    major, minor, patch = map(int, VERSION.fullmatch(latest).groups()) if latest else (0, 0, 0)
    patch += 1
    return f"v{major}.{minor}.{patch}"


if __name__ == "__main__":
    tag = plan(os.environ.get("RELEASE_VERSION", ""))
    with open(os.environ["GITHUB_OUTPUT"], "a") as output:
        output.write(f"tag={tag}\n")
    print(f"Release {tag}" if tag else "Only documentation changed; no release")
