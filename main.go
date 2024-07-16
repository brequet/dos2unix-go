package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: dos2unix <filename>")
		return
	}

	inputFileName := os.Args[1]
	outputFileName := generateOutputFileName(inputFileName)

	fmt.Printf("Reformatting file %s.. This may take a while\n", inputFileName)

	err := dos2unix(inputFileName, outputFileName)
	if err != nil {
		fmt.Printf("\nError: %v\n", err)
		return
	}

	fmt.Printf("\nFile %s successfully converted to UNIX EOL and saved as %s\n", inputFileName, outputFileName)
}

// generateOutputFileName generates the output file name by appending "formatted" before the file extension
func generateOutputFileName(inputFileName string) string {
	ext := filepath.Ext(inputFileName)
	base := strings.TrimSuffix(inputFileName, ext)
	return base + ".unix" + ext
}

func countLines(inputFileName string) (int, error) {
	fmt.Println("Checking file size..")
	file, err := os.Open(inputFileName)
	if err != nil {
		return 0, fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := reader.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

func dos2unix(inputFileName, outputFileName string) error {
	lineCount, err := countLines(inputFileName)
	if err != nil {
		return err
	}

	inputFile, err := os.Open(inputFileName)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outputFileName)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	reader := bufio.NewReader(inputFile)
	writer := bufio.NewWriter(outputFile)
	defer writer.Flush()

	progressIncrement := lineCount / 20 // 5% increments
	linesProcessed := 0
	progress := 0

	fmt.Printf("\rProgress: 0%%")

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				if len(line) > 0 {
					// Handle the case where the file does not end with a newline
					convertedLine := bytes.Replace(line, []byte("\r\n"), []byte("\n"), -1)
					writer.Write(convertedLine)
				}
				break
			}
			return fmt.Errorf("error reading file: %w", err)
		}

		convertedLine := bytes.Replace(line, []byte("\r\n"), []byte("\n"), -1)
		_, err = writer.Write(convertedLine)
		if err != nil {
			return fmt.Errorf("error writing to output file: %w", err)
		}

		linesProcessed++

		if linesProcessed%progressIncrement == 0 {
			progress += 5
			fmt.Printf("\rProgress: %d%%", progress)
		}
	}

	// Ensure we reach 100% progress
	fmt.Printf("\rProgress: 100%%")

	return nil
}
