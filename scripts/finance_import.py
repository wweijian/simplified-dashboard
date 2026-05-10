#!/usr/bin/env python3
"""Interactive finance CSV importer for simplified-dashboard."""

from __future__ import annotations

import argparse
import csv
import os
import shutil
import sqlite3
import sys
import termios
import tty
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Callable, Iterable


CREDIT_CARD_HEADER = [
    "Transaction Date",
    "Transaction Posting Date",
    "Transaction Description",
    "Transaction Type",
    "Payment Type",
    "Transaction Status",
    "Debit Amount",
    "Credit Amount",
]

DEBIT_HEADER = [
    "Transaction Date",
    "Transaction Code",
    "Description",
    "Transaction Ref1",
    "Transaction Ref2",
    "Transaction Ref3",
    "Status",
    "Debit Amount",
    "Credit Amount",
]

DATE_FORMATS = (
    "%Y-%m-%d",
    "%d/%m/%Y",
    "%m/%d/%Y",
    "%d-%m-%Y",
    "%m-%d-%Y",
    "%d %b %Y",
    "%d %B %Y",
)


@dataclass(frozen=True)
class StatementSpec:
    name: str
    header: list[str]
    date: str
    description: str
    debit: str
    credit: str


@dataclass(frozen=True)
class Category:
    id: int
    name: str
    type: str


@dataclass
class ImportRow:
    source_line: int
    date: str
    description: str
    amount: float
    category: Category | None


STATEMENT_SPECS = {
    "credit": StatementSpec(
        name="credit",
        header=CREDIT_CARD_HEADER,
        date="Transaction Date",
        description="Transaction Description",
        debit="Debit Amount",
        credit="Credit Amount",
    ),
    "debit": StatementSpec(
        name="debit",
        header=DEBIT_HEADER,
        date="Transaction Date",
        description="Description",
        debit="Debit Amount",
        credit="Credit Amount",
    ),
}


def main() -> int:
    return main_with_args()


def main_with_args(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    csv_path = Path(args.csv_file).expanduser()
    load_env_file(Path(args.env_file).expanduser())

    db_path_text = args.db or os.environ.get("DASHBOARD_DB_PATH", "")
    if db_path_text == "":
        print("DASHBOARD_DB_PATH is required in .env, or pass --db.", file=sys.stderr)
        return 1
    db_path = expand_path(db_path_text)

    if not csv_path.exists():
        print(f"CSV file not found: {csv_path}", file=sys.stderr)
        return 1
    if not db_path.exists():
        print(f"Database file not found: {db_path}", file=sys.stderr)
        return 1

    spec = choose_statement_spec(args.transaction_type)
    working_csv_path = prepare_working_csv(csv_path)

    try:
        raw_rows = read_statement_rows(working_csv_path, spec, args.encoding)
    except ValueError as err:
        print(err, file=sys.stderr)
        return 1

    with sqlite3.connect(db_path) as conn:
        categories = load_categories(conn)
        if not categories:
            print("No finance categories found. Open the dashboard once to seed categories.", file=sys.stderr)
            return 1

        non_interactive_category = resolve_non_interactive_category(args, categories)
        if args.yes and non_interactive_category is None:
            return 1
        stats = import_rows(
            conn,
            raw_rows,
            spec,
            categories,
            non_interactive_category,
            args.yes,
            consume_path=working_csv_path,
        )

    print_summary(stats)
    return 0


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Import credit/debit transactions into the dashboard DB.")
    parser.add_argument("csv_file", help="CSV file to import")
    parser.add_argument("--db", help="Path to dashboard SQLite database. Overrides DASHBOARD_DB_PATH from .env")
    parser.add_argument("--env-file", default=".env", help="Path to env file (default: .env)")
    parser.add_argument(
        "--transaction-type",
        choices=("credit", "credit-card", "debit"),
        help="Type of CSV statement. If omitted, the script asks.",
    )
    parser.add_argument("--category", help="Category to use with --yes imports")
    parser.add_argument("--yes", action="store_true", help="Accept every parsed row using --category")
    parser.add_argument("--encoding", default="utf-8-sig", help="CSV encoding (default: utf-8-sig)")
    return parser.parse_args(argv)


def prepare_working_csv(csv_path: Path) -> Path:
    working_path = dashboard_copy_path(csv_path)
    if working_path == csv_path:
        return csv_path
    if not working_path.exists():
        shutil.copy2(csv_path, working_path)
        print(f"Created resumable import copy: {working_path}")
    else:
        print(f"Resuming from import copy: {working_path}")
    return working_path


def dashboard_copy_path(csv_path: Path) -> Path:
    if csv_path.stem.endswith("_dashboard"):
        return csv_path
    return csv_path.with_name(f"{csv_path.stem}_dashboard{csv_path.suffix}")


def load_env_file(path: Path) -> None:
    if not path.exists():
        return

    with path.open(encoding="utf-8") as handle:
        for line_number, raw_line in enumerate(handle, start=1):
            line = raw_line.strip()
            if line == "" or line.startswith("#"):
                continue
            key, separator, value = line.partition("=")
            if separator == "":
                raise SystemExit(f"parse env file line {line_number}: expected KEY=VALUE")
            key = key.strip()
            if key == "":
                raise SystemExit(f"parse env file line {line_number}: empty key")
            if key in os.environ:
                continue
            os.environ[key] = os.path.expandvars(unquote(value.strip()))


def unquote(value: str) -> str:
    if len(value) < 2:
        return value
    if value[0] == '"' and value[-1] == '"':
        return value[1:-1].replace(r"\"", '"')
    if value[0] == "'" and value[-1] == "'":
        return value[1:-1]
    return value


def expand_path(value: str) -> Path:
    return Path(os.path.expandvars(value)).expanduser()


def choose_statement_spec(value: str | None) -> StatementSpec:
    if value:
        normalized = "credit" if value == "credit-card" else value
        return STATEMENT_SPECS[normalized]

    print("What type of transactions are you importing?")
    print("  1. Credit card")
    print("  2. Debit")
    while True:
        choice = read_key("> ").lower()
        if choice in {"1", "c"}:
            return STATEMENT_SPECS["credit"]
        if choice in {"2", "d"}:
            return STATEMENT_SPECS["debit"]
        print("Choose 1 for credit card or 2 for debit.")


def read_statement_rows(csv_path: Path, spec: StatementSpec, encoding: str) -> list[tuple[int, dict[str, str]]]:
    with csv_path.open(newline="", encoding=encoding) as handle:
        reader = csv.reader(handle)
        header_line = 0
        for line_number, row in enumerate(reader, start=1):
            if row_matches_header(row, spec.header):
                header_line = line_number
                break

        if header_line == 0:
            raise ValueError(f"Could not find {spec.name} transaction header in {csv_path}")

        rows = []
        for line_number, row in enumerate(reader, start=header_line + 1):
            if row_is_empty(row):
                continue
            padded = row + [""] * max(len(spec.header) - len(row), 0)
            rows.append((line_number, dict(zip(spec.header, padded))))
        return rows


def row_matches_header(row: list[str], expected: list[str]) -> bool:
    cells = [clean_cell(cell) for cell in row]
    return len(cells) >= len(expected) and cells[: len(expected)] == expected


def clean_cell(value: str) -> str:
    return value.replace("\ufeff", "").strip()


def row_is_empty(row: Iterable[str]) -> bool:
    return all(clean_cell(cell) == "" for cell in row)


def load_categories(conn: sqlite3.Connection) -> list[Category]:
    rows = conn.execute(
        """
        SELECT id, name, type
        FROM finance_categories
        ORDER BY type ASC, name ASC
        """
    ).fetchall()
    return [Category(id=row[0], name=row[1], type=row[2]) for row in rows]


def find_category(categories: list[Category], name: str) -> Category | None:
    normalized = name.strip().lower()
    for category in categories:
        if category.name.lower() == normalized:
            return category
    return None


def resolve_non_interactive_category(args: argparse.Namespace, categories: list[Category]) -> Category | None:
    if not args.yes:
        return None
    if not args.category:
        print("--yes requires --category because statement rows do not include categories.", file=sys.stderr)
        return None

    category = find_category(categories, args.category)
    if category is None:
        print(f"Category not found: {args.category}", file=sys.stderr)
        return None
    return category


def import_rows(
    conn: sqlite3.Connection,
    raw_rows: Iterable[tuple[int, dict[str, str]]],
    spec: StatementSpec,
    categories: list[Category],
    non_interactive_category: Category | None,
    accept_all: bool,
    consume_path: Path | None = None,
) -> dict[str, int]:
    stats = {"imported": 0, "skipped": 0, "duplicates": 0, "errors": 0}
    consumed_lines = 0

    for source_line, raw in raw_rows:
        try:
            item = parse_import_row(source_line, raw, spec)
        except ValueError as err:
            stats["errors"] += 1
            print(f"Line {source_line}: {err}")
            continue

        duplicate = is_duplicate(conn, item)
        if duplicate:
            stats["duplicates"] += 1

        if accept_all:
            item.category = non_interactive_category
            insert_transaction(conn, item)
            consumed_lines += consume_accepted_line(consume_path, source_line, consumed_lines)
            stats["imported"] += 1
            continue

        action = review_row(conn, item, categories, duplicate)
        if action == "quit":
            break
        if action == "skip":
            stats["skipped"] += 1
            continue

        insert_transaction(conn, item)
        consumed_lines += consume_accepted_line(consume_path, source_line, consumed_lines)
        stats["imported"] += 1

    conn.commit()
    return stats


def consume_accepted_line(path: Path | None, source_line: int, consumed_lines: int) -> int:
    if path is None:
        return 0
    remove_line(path, source_line - consumed_lines)
    return 1


def remove_line(path: Path, line_number: int) -> None:
    lines = path.read_text(encoding="utf-8-sig").splitlines(keepends=True)
    if line_number < 1 or line_number > len(lines):
        raise ValueError(f"cannot remove line {line_number} from {path}")
    del lines[line_number - 1]
    path.write_text("".join(lines), encoding="utf-8")


def parse_import_row(source_line: int, raw: dict[str, str], spec: StatementSpec) -> ImportRow:
    date = parse_date(raw.get(spec.date, ""))
    description = clean_cell(raw.get(spec.description, ""))
    amount = parse_debit_credit_amount(raw.get(spec.debit, ""), raw.get(spec.credit, ""))
    return ImportRow(source_line=source_line, date=date, description=description, amount=amount, category=None)


def parse_date(value: str) -> str:
    value = clean_cell(value)
    for date_format in DATE_FORMATS:
        try:
            return datetime.strptime(value, date_format).date().isoformat()
        except ValueError:
            pass
    raise ValueError(f"unsupported date {value!r}")


def parse_debit_credit_amount(debit_text: str, credit_text: str) -> float:
    debit = parse_optional_amount(debit_text)
    credit = parse_optional_amount(credit_text)
    amount = credit - debit
    if amount == 0:
        raise ValueError("missing amount")
    return amount


def parse_optional_amount(value: str) -> float:
    value = clean_cell(value)
    if value == "":
        return 0
    return abs(parse_amount(value))


def parse_amount(value: str) -> float:
    value = clean_cell(value)
    if not value:
        raise ValueError("missing amount")

    negative = value.startswith("(") and value.endswith(")")
    cleaned = value.replace("$", "").replace(",", "").replace(" ", "").strip("()")
    amount = float(cleaned)
    if negative:
        amount = -amount
    return amount


def review_row(conn: sqlite3.Connection, item: ImportRow, categories: list[Category], duplicate: bool) -> str:
    clear_screen()
    print_review_row(item, duplicate)
    while True:
        action = read_key("[a]ccept  [e]dit  [s]kip  [q]uit > ").lower()
        if action in {"\r", "\n", "a"}:
            item.category = ensure_category(item.category, categories)
            return "accept"
        if action == "s":
            return "skip"
        if action == "q":
            return "quit"
        if action == "e":
            edit_row(item, categories)
            duplicate = is_duplicate(conn, item)


def print_review_row(item: ImportRow, duplicate: bool) -> None:
    print()
    print(f"Line {item.source_line}:")
    print(f"  Date:        {item.date}")
    print(f"  Description: {item.description}")
    print(f"  Amount:      {format_amount(item.amount)}")
    print(f"  Category:    {category_label(item.category)}")
    if duplicate:
        print("  Warning: duplicate date + description + amount already exists")


def clear_screen() -> None:
    if sys.stdout.isatty():
        print("\033[2J\033[H", end="")


def edit_row(item: ImportRow, categories: list[Category]) -> None:
    item.date = prompt_value("Date", item.date, parse_date)
    item.description = prompt_value("Description", item.description, lambda value: value.strip())
    item.amount = prompt_value("Amount", item.amount, parse_amount)
    item.category = prompt_category(categories, item.category)


def category_label(category: Category | None) -> str:
    if category is None:
        return "(required before accept)"
    return category.name


def ensure_category(category: Category | None, categories: list[Category]) -> Category:
    if category is not None:
        return category
    return prompt_category(categories, None)


def prompt_value(label: str, current: object, parser: Callable[[str], object]):
    while True:
        value = input(f"{label} [{current}]: ").strip()
        if value == "":
            return current
        try:
            return parser(value)
        except ValueError as err:
            print(err)


def prompt_category(categories: list[Category], current: Category | None) -> Category:
    shortcuts = category_shortcuts(categories)
    print("Categories:")
    for category in categories:
        marker = "*" if current is not None and category.id == current.id else " "
        print(f"  {marker} {shortcuts[category.id]}. {category.name} ({category.type})")

    while True:
        prompt = f"Category [{current.name}] > " if current is not None else "Category > "
        value = read_key(prompt).lower()
        if value in {"\r", "\n"} and current is not None:
            return current
        if value in {"\r", "\n"}:
            print("Category is required.")
            continue
        category = category_by_shortcut(categories, shortcuts, value)
        if category is not None:
            return category
        print("Choose one of the category shortcut keys.")


def read_key(prompt: str) -> str:
    print(prompt, end="", flush=True)
    if not sys.stdin.isatty():
        value = input().strip()
        print()
        return value[:1] if value else "\n"

    fd = sys.stdin.fileno()
    old_settings = termios.tcgetattr(fd)
    try:
        tty.setraw(fd)
        value = sys.stdin.read(1)
    finally:
        termios.tcsetattr(fd, termios.TCSADRAIN, old_settings)
    print(value)
    return value


def category_shortcuts(categories: list[Category]) -> dict[int, str]:
    keys = list("1234567890abcdefghijklmnopqrstuvwxyz")
    if len(categories) > len(keys):
        raise ValueError(f"too many categories for single-key selection; max is {len(keys)}")
    return {category.id: keys[index] for index, category in enumerate(categories)}


def category_by_shortcut(
    categories: list[Category],
    shortcuts: dict[int, str],
    key: str,
) -> Category | None:
    for category in categories:
        if shortcuts[category.id] == key:
            return category
    return None


def is_duplicate(conn: sqlite3.Connection, item: ImportRow) -> bool:
    row = conn.execute(
        """
        SELECT 1
        FROM finance_transactions
        WHERE date = ? AND COALESCE(description, '') = ? AND amount = ?
        LIMIT 1
        """,
        (item.date, item.description, item.amount),
    ).fetchone()
    return row is not None


def insert_transaction(conn: sqlite3.Connection, item: ImportRow) -> None:
    if item.category is None:
        raise ValueError("category is required before inserting transaction")
    conn.execute(
        """
        INSERT INTO finance_transactions (date, amount, category_id, description)
        VALUES (?, ?, ?, ?)
        """,
        (item.date, item.amount, item.category.id, item.description or None),
    )


def format_amount(amount: float) -> str:
    sign = "-" if amount < 0 else ""
    return f"{sign}${abs(amount):,.2f}"


def print_summary(stats: dict[str, int]) -> None:
    print()
    print("Import summary:")
    print(f"  Imported:   {stats['imported']}")
    print(f"  Skipped:    {stats['skipped']}")
    print(f"  Duplicates: {stats['duplicates']}")
    print(f"  Errors:     {stats['errors']}")


if __name__ == "__main__":
    raise SystemExit(main())
