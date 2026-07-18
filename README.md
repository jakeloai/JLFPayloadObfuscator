# JLFPayloadObfuscator

> Designed by **jakelo.ai** · Coded with AI assistance

You give it a payload string. It gives you back multiple encoded and mutated versions of that same string.

---

## What it does

You have a payload — maybe a SQL injection string, an XSS vector, or a PHP command. You need to test if a WAF or input filter blocks it. This tool generates variants:

- URL encoded, double URL encoded, hex encoded
- HTML entity encoded, Unicode escaped, base64 encoded
- Case-randomized keywords (for SQL and XSS)
- PHP-specific wrappers (base64 eval, XOR dynamic strings, bitwise NOT)

You feed the output into ffuf, Burp Intruder, or any fuzzer. It does not scan, attack, or connect anywhere. It just transforms strings.

---

## Installation

```bash
git clone https://github.com/jakeloai/JLFPayloadObfuscator.git
cd JLFPayloadObfuscator
go build -o jlfpayload .
```

Requires Go 1.18+.

---

## How to use

**Single payload:**
```bash
./jlfpayload -p "' OR 1=1--" -m sql -o sql_wordlist.txt
```

**Multiple payloads from file:**
```bash
./jlfpayload -f payloads.txt -m all -o obfuscated_wordlist.txt
```

**With ffuf:**
```bash
./jlfpayload -f my_payloads.txt -m all -o fuzz_wordlist.txt
ffuf -w fuzz_wordlist.txt -u "https://target.com/api?param=FUZZ"
```

---

## Modes

| Mode | What it generates |
|---|---|
| `general` | URL, hex, HTML entity, Unicode, base64, and chained variants |
| `sql` | General encodings + comment substitution (`/**/`, `+`) + keyword case randomization |
| `xss` | General encodings + tag/event case randomization + `javascript:` protocol mutation |
| `php` | General encodings + base64 eval wrappers + XOR/NOT dynamic string generation, with PHP version awareness |
| `all` | Everything above |

---

## PHP version handling

Some PHP functions were removed or changed across versions. The tool targets the right mutation for the version you are testing:

```bash
./jlfpayload -p "<?php system($_GET['cmd']); ?>" -m php -phpv legacy   # PHP 5.x
./jlfpayload -p "<?php system($_GET['cmd']); ?>" -m php -phpv modern   # PHP 7.x
./jlfpayload -p "<?php system($_GET['cmd']); ?>" -m php -phpv latest    # PHP 8.x
```

---

## Options

| Flag | Description |
|---|---|
| `-p` | Single payload string |
| `-f` | File with one payload per line |
| `-o` | Output file (default: `obfuscated_wordlist.txt`) |
| `-m` | Mode: `sql`, `xss`, `php`, `general`, `all` (default: `all`) |
| `-phpv` | PHP version target: `legacy`, `modern`, `latest` (default: `modern`) |
| `-q` | Quiet mode |

---

## What it does not do

- It does not parse SQL/HTML/PHP syntax trees
- It does not perform active scanning or vulnerability detection
- It does not handle binary protocols
- It does not generate novel exploits — only mutates what you give it
- It does not bypass behavioral rules like rate limiting or IP reputation

It is a preprocessor. You bring the payload, it brings the variants.

---

## License

MIT © jakelo.ai
