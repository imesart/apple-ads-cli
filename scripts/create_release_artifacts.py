#!/usr/bin/env python3

import argparse
import hashlib
import os
import pathlib
import re
import shutil
import subprocess
import sys
import tarfile
import zipfile


TARGETS = [
    ("darwin", "amd64"),
    ("darwin", "arm64"),
    ("linux", "amd64"),
    ("linux", "arm64"),
    ("windows", "amd64"),
    ("windows", "arm64"),
]

SEMVER_RE = re.compile(r"^\d+\.\d+\.\d+$")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Build release archives for aads.")
    parser.add_argument("--version", required=True, help="Release version in x.y.z form")
    parser.add_argument("--commit", default="unknown", help="Git commit to embed")
    parser.add_argument("--date", default="unknown", help="Build date to embed")
    parser.add_argument("--output-dir", default="release", help="Directory for archives")
    parser.add_argument("--source-dir", default=".", help="Go package/module to build")
    parser.add_argument("--binary-name", default="aads", help="Binary name")
    return parser.parse_args()


def validate_version(version: str) -> None:
    if not SEMVER_RE.fullmatch(version):
        raise ValueError(f'invalid version "{version}": expected x.y.z')


def archive_name(binary_name: str, version: str, goos: str, goarch: str) -> str:
    ext = ".zip" if goos == "windows" else ".tar.gz"
    return f"{binary_name}_{version}_{goos}_{goarch}{ext}"


def binary_file_name(binary_name: str, goos: str) -> str:
    return binary_name + ".exe" if goos == "windows" else binary_name


def build_target(args: argparse.Namespace, output_path: pathlib.Path, goos: str, goarch: str, gocache: pathlib.Path) -> None:
    ldflags = f"-X main.version={args.version} -X main.commit={args.commit} -X main.date={args.date}"
    env = os.environ.copy()
    env["CGO_ENABLED"] = "0"
    env["GOOS"] = goos
    env["GOARCH"] = goarch
    env["GOCACHE"] = str(gocache)
    subprocess.run(
        ["go", "build", "-ldflags", ldflags, "-o", str(output_path), args.source_dir],
        check=True,
        env=env,
    )


def sign_target(binary_path: pathlib.Path, goos: str, developer_id: str) -> None:
    if goos != "darwin":
        return
    if not developer_id:
        print(
            f"warning: APPLE_DEVELOPER_ID not set, skipping codesign for {binary_path.name}",
            file=sys.stderr,
        )
        return
    if sys.platform != "darwin":
        raise RuntimeError(f"codesign requested for {binary_path.name} but host OS is {sys.platform}")
    subprocess.run(
        [
            "codesign",
            "--force",
            "--sign",
            developer_id,
            "--timestamp",
            "--options=runtime",
            str(binary_path),
        ],
        check=True,
    )


def create_archive(binary_path: pathlib.Path, archive_path: pathlib.Path, archive_binary_name: str, goos: str) -> None:
    if goos == "windows":
        with zipfile.ZipFile(archive_path, "w", compression=zipfile.ZIP_DEFLATED) as archive:
            archive.write(binary_path, arcname=archive_binary_name)
        return

    with tarfile.open(archive_path, "w:gz") as archive:
        archive.add(binary_path, arcname=archive_binary_name)


def file_sha256(path: pathlib.Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def main() -> int:
    args = parse_args()
    try:
        validate_version(args.version)
    except ValueError as exc:
        print(f"Error: {exc}", file=sys.stderr)
        return 2

    output_dir = pathlib.Path(args.output_dir).resolve()
    build_dir = output_dir / ".build"
    gocache_dir = output_dir / ".gocache"

    shutil.rmtree(output_dir, ignore_errors=True)
    build_dir.mkdir(parents=True, exist_ok=True)
    gocache_dir.mkdir(parents=True, exist_ok=True)

    developer_id = os.environ.get("APPLE_DEVELOPER_ID", "")
    archives: list[tuple[pathlib.Path, str]] = []

    for goos, goarch in TARGETS:
        binary_path = build_dir / f"{args.binary_name}_{goos}_{goarch}"
        if goos == "windows":
            binary_path = binary_path.with_suffix(".exe")

        try:
            build_target(args, binary_path, goos, goarch, gocache_dir)
            sign_target(binary_path, goos, developer_id)
        except subprocess.CalledProcessError as exc:
            print(f"Error: build {goos}/{goarch}: {exc}", file=sys.stderr)
            return 1
        except RuntimeError as exc:
            print(f"Error: {exc}", file=sys.stderr)
            return 1

        archive_path = output_dir / archive_name(args.binary_name, args.version, goos, goarch)
        create_archive(binary_path, archive_path, binary_file_name(args.binary_name, goos), goos)
        archives.append((archive_path, file_sha256(archive_path)))

    archives.sort(key=lambda item: item[0].name)
    checksum_path = output_dir / f"{args.binary_name}_{args.version}_checksums.txt"
    with checksum_path.open("w", encoding="utf-8") as handle:
        for archive_path, digest in archives:
            handle.write(f"{digest}  {archive_path.name}\n")

    print("Generated release archives:")
    for archive_path, _ in archives:
        print(f"- {archive_path}")
    print(f"Checksums: {checksum_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
