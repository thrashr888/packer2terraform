package packer2terraform

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/template"
)

// LogLine encapsulates a single log line from the csv
type LogLine struct {
	timestamp     string
	builderTarget string
	lineType      string
	messageType   string
	messageTypeI  int
	messageA      string
	messageB      string
}

// NewLogLine constructs a LogLine
func NewLogLine(v []string) *LogLine {
	l := &LogLine{"", "", "", "", 0, "", ""}

	if len(v) > 0 {
		l.timestamp = v[0]
	}
	if len(v) > 1 {
		l.builderTarget = v[1]
	}
	if len(v) > 2 {
		l.lineType = v[2]
	}
	if len(v) > 3 {
		l.messageType = v[3]
	}
	if len(v) > 4 {
		l.messageA = v[4]
	}
	if len(v) > 5 {
		l.messageB = v[5]
	}
	if len(l.messageType) > 0 {
		l.messageTypeI, _ = strconv.Atoi(l.messageType)
	}
	return l
}

// Artifact is our representation of a Packer.Artifact
type Artifact struct {
	BuilderTarget string
	BuilderID     string
	ID            string
	IDSplit       []string
	Message       string
	FilesCount    string
}

// ApplyLogLine uses a LogLine to
func (a *Artifact) ApplyLogLine(line LogLine) {
	if line.messageA == "builder-id" {
		a.BuilderID = line.messageB
	}
	if line.messageA == "id" {
		a.ID = line.messageB
		a.IDSplit = strings.Split(line.messageB, ":")
	}
	if line.messageA == "string" {
		a.Message = line.messageB
	}
	if line.messageA == "files-count" {
		a.FilesCount = line.messageB
	}
	if line.messageA == "nil" {
		// no file
	}
}

// From the Packer docs, this represents:
// 1 index, 2 subtype, 3..n subtype data

type templatePage struct {
	Artifacts []Artifact
}

// ErrMissing when artifact-count is higher than the Artifacts found
type ErrMissing struct {
	count int
}

func (e *ErrMissing) Error() string {
	return fmt.Sprintf("Missing %d artifacts.", e.count)
}

// ErrList when there's errors mentioned in the CSV data
type ErrList struct {
	List []string
}

func (e *ErrList) Error() string {
	return "List of errors: " + strings.Join(e.List, "; ")
}

// Add an error string to the list of errors
func (e *ErrList) Add(err string) {
	e.List = append(e.List, err)
}

// ErrNotFound when there's no artifacts found in the CSV data
var ErrNotFound = errors.New("No Artifacts found.")

// A simple terraform template for aws amis in zones
var TemplateAmazonEBS = `variable "images" {
    default = {
{{range .Artifacts}}
        {{index .IDSplit 0}} = "{{index .IDSplit 1}}"{{end}}
    }
}`

// ReadCSV converts the csv files into a data structure we can use
func ReadCSV(csvReader io.Reader) (ret [][]string, err error) {
	reader := csv.NewReader(csvReader)
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	return reader.ReadAll()
}

// ExtractArtifacts extracts Artifacts from array of LogLines
func ExtractArtifacts(parsed [][]string) (artifacts []Artifact, err error) {
	var errorCount int
	var errorMsg ErrList
	var artifactCount int

	for _, v := range parsed {
		if len(v) < 2 {
			// Not enough data to be useful
			continue
		}

		// Build a LogLine
		line := NewLogLine(v)

		// Artifacts:
		if line.lineType == "artifact-count" {
			artifactCount = line.messageTypeI
		}
		if line.lineType == "artifact" {
			if len(artifacts) < line.messageTypeI+1 {
				a := Artifact{}
				a.BuilderTarget = line.builderTarget
				artifacts = append(artifacts, a)
			}

			a := &artifacts[line.messageTypeI]
			a.ApplyLogLine(*line)
		}

		// Errors:
		if line.lineType == "error-count" && line.messageTypeI > 0 {
			errorCount = line.messageTypeI
		}
		if line.lineType == "error" {
			errorMsg.Add(line.messageType)
		}
	}

	if artifactCount < len(artifacts) {
		artifactsMissing := artifactCount - len(artifacts)
		return nil, &ErrMissing{artifactsMissing}
	}

	if errorCount > 0 && len(errorMsg.List) > 0 {
		return nil, &errorMsg
	}

	// Clean up empty artifacts
	for i, artifact := range artifacts {
		if artifact.ID == "" {
			artifacts = append(artifacts[:i], artifacts[i+1:]...)
		}
	}
	if len(artifacts) == 0 {
		return nil, ErrNotFound
	}

	return artifacts, nil
}

// ToTemplate applies the artifacts to a given template string
func ToTemplate(artifacts []Artifact, tmpl string) (ret string, err error) {
	// Setup the page vars
	var thePage = templatePage{}
	thePage.Artifacts = artifacts

	t := template.Must(template.New("tmpl").Parse(tmpl))

	var doc bytes.Buffer
	t.Execute(&doc, thePage)
	ret = doc.String()

	return ret, nil
}
