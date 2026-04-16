#!/usr/bin/env python3

from __future__ import annotations

import argparse
import re
import subprocess
import sys
from dataclasses import dataclass, field
from pathlib import Path


SECTION_RE = re.compile(r"^[A-Z][A-Z ]+$")
SUBCOMMAND_RE = re.compile(r"^\s{2,}(\S+)\s{2,}(.+?)\s*$")


@dataclass
class CommandDoc:
    path: list[str]
    description: str
    usage: str = ""
    subcommands_block: str = ""
    flags: str = ""
    children: list["CommandDoc"] = field(default_factory=list)

    @property
    def is_root(self) -> bool:
        return not self.path

    @property
    def heading(self) -> str:
        return " ".join(self.path)

    @property
    def depth(self) -> int:
        return len(self.path)


def parse_sections(text: str) -> dict[str, str]:
    sections: dict[str, list[str]] = {}
    current: str | None = None

    for raw_line in text.splitlines():
        line = raw_line.rstrip()
        stripped = line.strip()
        if SECTION_RE.fullmatch(stripped):
            current = stripped
            sections[current] = []
            continue
        if current is not None:
            sections[current].append(line)

    return {name: trim_block(lines) for name, lines in sections.items()}


def trim_block(lines: list[str]) -> str:
    start = 0
    end = len(lines)

    while start < end and not lines[start].strip():
        start += 1
    while end > start and not lines[end - 1].strip():
        end -= 1

    return "\n".join(lines[start:end]).rstrip()


def normalize_paragraphs(block: str) -> str:
    if not block:
        return ""

    paragraphs: list[str] = []
    current: list[str] = []

    for line in block.splitlines():
        stripped = line.strip()
        if not stripped:
            if current:
                paragraphs.append(" ".join(current))
                current = []
            continue
        current.append(stripped)

    if current:
        paragraphs.append(" ".join(current))

    return "\n\n".join(paragraphs)


def run_help(repo_root: Path, path: list[str]) -> str:
    cmd = [str(repo_root / "bin" / "aads"), "help", *path]
    result = subprocess.run(
        cmd,
        cwd=repo_root,
        capture_output=True,
        text=True,
        check=False,
    )
    if result.returncode != 0:
        raise RuntimeError(
            f"command failed ({result.returncode}): {' '.join(cmd)}\n{result.stderr.strip()}"
        )
    return result.stdout.strip()


def build_tree(repo_root: Path, path: list[str]) -> CommandDoc:
    text = run_help(repo_root, path)
    sections = parse_sections(text)
    description = normalize_paragraphs(sections.get("DESCRIPTION", ""))

    node = CommandDoc(
        path=path,
        description=description,
        usage=sections.get("USAGE", ""),
        subcommands_block=sections.get("SUBCOMMANDS", ""),
        flags=sections.get("FLAGS", ""),
    )

    for name in parse_subcommands(node.subcommands_block):
        node.children.append(build_tree(repo_root, [*path, name]))

    return node


def parse_subcommands(block: str) -> list[str]:
    names: list[str] = []
    for line in block.splitlines():
        match = SUBCOMMAND_RE.match(line)
        if match:
            names.append(match.group(1))
    return names


def fenced(block: str) -> str:
    return f"```\n{block.rstrip()}\n```"


def join_sections(sections: list[tuple[str, str]]) -> str:
    blocks: list[str] = []
    for title, content in sections:
        if not content:
            continue
        blocks.append(f"{title}\n{content}")
    return "\n\n".join(blocks)


def render_node(node: CommandDoc) -> str:
    if node.is_root:
        return render_root(node)

    parts: list[str] = []
    heading_level = "#" * (node.depth + 1)
    parts.append(f"{heading_level} {node.heading}")
    parts.append("")

    if node.description:
        parts.append(node.description)
        parts.append("")

    main_block = join_sections(
        [
            ("USAGE", node.usage),
            ("SUBCOMMANDS", node.subcommands_block),
        ]
    )
    if main_block:
        parts.append(fenced(main_block))
        parts.append("")

    if node.flags:
        parts.append(fenced(f"FLAGS\n{node.flags}"))
        parts.append("")

    child_blocks = [render_node(child) for child in node.children]
    if child_blocks:
        parts.append("\n\n".join(child_blocks))

    return "\n".join(parts).rstrip()


def render_root(node: CommandDoc) -> str:
    parts = [
        "# aads Command Reference",
        "",
        "> Generated with `make commands-doc`. Do not edit manually.",
        "",
    ]

    if node.description:
        parts.extend([node.description, ""])

    root_block = join_sections(
        [
            ("USAGE", node.usage),
            ("SUBCOMMANDS", node.subcommands_block),
        ]
    )
    if root_block:
        parts.extend([fenced(root_block), ""])

    if node.flags:
        parts.extend(["## Global Flags", "", fenced(f"FLAGS\n{node.flags}"), ""])

    for child in node.children:
        parts.extend(["---", "", render_node(child), ""])

    return "\n".join(parts).rstrip() + "\n"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Generate docs/commands.md from CLI help output."
    )
    parser.add_argument("output_path", help="Path to the generated markdown file.")
    parser.add_argument(
        "--check",
        action="store_true",
        help="Check whether the output file is up to date instead of writing it.",
    )
    return parser.parse_args()


def main() -> None:
    repo_root = Path(__file__).resolve().parent.parent
    args = parse_args()
    output_path = Path(args.output_path)
    if not output_path.is_absolute():
        output_path = repo_root / output_path
    tree = build_tree(repo_root, [])
    rendered = render_root(tree)

    if args.check:
        current = output_path.read_text(encoding="utf-8") if output_path.exists() else ""
        if current != rendered:
            print(
                f"{output_path} is out of date; run `make commands-doc`.",
                file=sys.stderr,
            )
            raise SystemExit(1)
        return

    output_path.write_text(rendered, encoding="utf-8")


if __name__ == "__main__":
    main()
