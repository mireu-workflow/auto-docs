package agent

import (
	"testing"
)

func TestCleanReportOutput(t *testing.T) {
	inputWithPreamble := `Junior Engineer reviewing code to create a Wiki (User Guide and Implementation Spec)...
* Project Name: Auth-System
* Tech Stack: Go

# 🚀 Auth-System User Guide & Technical Spec

## 📌 Key Features
- Multi-DB support`

	expected := `# 🚀 Auth-System User Guide & Technical Spec

## 📌 Key Features
- Multi-DB support`

	result := cleanReportOutput(inputWithPreamble)
	if result != expected {
		t.Errorf("expected:\n%s\n\ngot:\n%s", expected, result)
	}

	inputWithCodeblock := "```markdown\n# 🚀 Title\nContent\n```"
	expectedCodeblock := "# 🚀 Title\nContent"

	resultCodeblock := cleanReportOutput(inputWithCodeblock)
	if resultCodeblock != expectedCodeblock {
		t.Errorf("expected:\n%s\n\ngot:\n%s", expectedCodeblock, resultCodeblock)
	}
}
