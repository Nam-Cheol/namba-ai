import hashlib
import os
import shutil
import subprocess
import tarfile
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "install.sh"
ASSET = "namba_Linux_x86_64.tar.gz"


class InstallShTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)
        self.install_dir = self.root / "bin"
        self.home = self.root / "home"
        self.home.mkdir()

    def tearDown(self):
        self.tmp.cleanup()

    def write_archive(self, name="namba", content=b"namba test binary"):
        archive = self.root / ASSET
        source_dir = Path(tempfile.mkdtemp(dir=self.root))
        if name:
            binary = source_dir / name
            binary.write_bytes(content)
            binary.chmod(0o755)
        with tarfile.open(archive, "w:gz") as tar:
            if name:
                tar.add(source_dir / name, arcname=name)
        return archive

    def write_checksums(self, archive, lines):
        checksums = self.root / "checksums.txt"
        digest = hashlib.sha256(archive.read_bytes()).hexdigest()
        checksums.write_text("\n".join(line.format(digest=digest) for line in lines) + "\n")
        return checksums

    def run_installer(self, archive, checksums):
        env = os.environ.copy()
        env.update(
            {
                "HOME": str(self.home),
                "NAMBA_INSTALL_TEST_ASSET_PATH": str(archive),
                "NAMBA_INSTALL_TEST_CHECKSUMS_PATH": str(checksums),
                "NAMBA_INSTALL_TEST_DIR": str(self.install_dir),
                "NAMBA_INSTALL_TEST_VERSION": "v0.0.0-test",
                "NAMBA_INSTALL_TEST_OS": "Linux",
                "NAMBA_INSTALL_TEST_ARCH": "x86_64",
            }
        )
        return subprocess.run(
            ["sh", str(SCRIPT)],
            cwd=ROOT,
            env=env,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
        )

    def assert_installed(self):
        installed = self.install_dir / "namba"
        self.assertTrue(installed.exists(), "expected namba to be installed")
        self.assertEqual(installed.read_bytes(), b"namba test binary")
        self.assertFalse((self.home / ".profile").exists())
        self.assertFalse((self.home / ".zshrc").exists())

    def test_installs_after_checksum_verification_for_supported_formats(self):
        formats = [
            "{digest}  " + ASSET,
            "{digest} *" + ASSET,
            "{digest}  ./" + ASSET,
        ]
        for checksum_line in formats:
            with self.subTest(checksum_line=checksum_line):
                if self.install_dir.exists():
                    shutil.rmtree(self.install_dir)
                archive = self.write_archive()
                checksums = self.write_checksums(archive, [checksum_line])

                result = self.run_installer(archive, checksums)

                self.assertEqual(result.returncode, 0, result.stderr)
                self.assert_installed()

    def test_checksum_mismatch_fails_before_install(self):
        archive = self.write_archive()
        checksums = self.root / "checksums.txt"
        checksums.write_text("0" * 64 + "  " + ASSET + "\n")

        result = self.run_installer(archive, checksums)

        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Checksum verification failed", result.stderr)
        self.assertFalse((self.install_dir / "namba").exists())

    def test_missing_checksum_line_fails_closed(self):
        archive = self.write_archive()
        checksums = self.write_checksums(archive, ["{digest}  other.tar.gz"])

        result = self.run_installer(archive, checksums)

        self.assertNotEqual(result.returncode, 0)
        self.assertIn("does not contain an entry", result.stderr)
        self.assertFalse((self.install_dir / "namba").exists())

    def test_checksum_matching_is_exact_by_asset_basename(self):
        archive = self.write_archive()
        checksums = self.write_checksums(
            archive,
            [
                "0" * 64 + "  " + ASSET + ".bak",
                "{digest}  ./subdir/" + ASSET,
            ],
        )

        result = self.run_installer(archive, checksums)

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assert_installed()

    def test_missing_expected_binary_fails_after_verified_extraction(self):
        archive = self.write_archive(name="not-namba")
        checksums = self.write_checksums(archive, ["{digest}  " + ASSET])

        result = self.run_installer(archive, checksums)

        self.assertNotEqual(result.returncode, 0)
        self.assertIn("namba was not found", result.stderr)
        self.assertFalse((self.install_dir / "namba").exists())


if __name__ == "__main__":
    unittest.main()
