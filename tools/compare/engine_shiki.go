package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type shikiInput struct {
	Name string `json:"name"`
	Lang string `json:"lang"`
	Code string `json:"code"`
}

type shikiOutput struct {
	Name   string  `json:"name"`
	Lang   string  `json:"lang"`
	ColdMs float64 `json:"coldMs"`
	WarmMs float64 `json:"warmMs"`
	Tokens int     `json:"tokens"`
	Scopes int     `json:"scopes"`
	Dump   string  `json:"dump"`
}

func benchShiki(inputs []Input, iters int, theme string) (map[string]EngineResult, map[string]string, error) {
	scriptDir, err := findScriptDir()
	if err != nil {
		return nil, nil, err
	}

	si := make([]shikiInput, len(inputs))
	for i, inp := range inputs {
		si[i] = shikiInput{Name: inp.Name, Lang: inp.Lang, Code: inp.Code}
	}
	tmpFile, err := os.CreateTemp("", "shiki-inputs-*.json")
	if err != nil {
		return nil, nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	if err := json.NewEncoder(tmpFile).Encode(si); err != nil {
		tmpFile.Close()
		return nil, nil, fmt.Errorf("write inputs: %w", err)
	}
	tmpFile.Close()

	cmd := exec.Command("node", filepath.Join(scriptDir, "shiki_bench.mjs"),
		tmpFile.Name(), fmt.Sprintf("%d", iters), theme)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, nil, fmt.Errorf("node shiki_bench.mjs: %w", err)
	}

	results := make(map[string]EngineResult, len(inputs))
	dumps := make(map[string]string, len(inputs))
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var so shikiOutput
		if err := json.Unmarshal([]byte(line), &so); err != nil {
			return nil, nil, fmt.Errorf("parse shiki output: %w\nline: %s", err, line)
		}
		results[so.Name] = EngineResult{
			ColdMs: so.ColdMs,
			WarmMs: so.WarmMs,
			Tokens: so.Tokens,
			Scopes: so.Scopes,
		}
		if so.Dump != "" {
			dumps[so.Lang] = so.Dump
		}
	}
	return results, dumps, nil
}

func findScriptDir() (string, error) {
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exe)
		if _, err := os.Stat(filepath.Join(dir, "shiki_bench.mjs")); err == nil {
			return dir, nil
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(wd, "shiki_bench.mjs")); err == nil {
		return wd, nil
	}
	return "", fmt.Errorf("cannot find shiki_bench.mjs in executable dir or working dir")
}
