import argparse
from pathlib import Path

from app.converter import convert

def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Convert a Markdown file to HTML using the GitHub Markdown API."
    )
    parser.add_argument("input", help="Path to the Markdown file to convert")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    input_path = Path(args.input).expanduser().resolve()
    if not input_path.exists():
        raise FileNotFoundError(f"Input file not found: {input_path}")

    markdown_text = input_path.read_text(encoding="utf-8")
    output_text = convert(markdown_text, fallback_title=input_path.stem)

    output_path = input_path.with_suffix(".html")
    output_path.write_text(output_text, encoding="utf-8")

    print(f"Saved: {output_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
