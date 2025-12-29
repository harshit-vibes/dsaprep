package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	v1 "github.com/harshit-vibes/cf/pkg/internal/schema/v1"
	"gopkg.in/yaml.v3"
)

// SaveProblem saves a problem to the workspace
func (w *Workspace) SaveProblem(problem *v1.Problem) error {
	problemDir := w.ProblemPath(problem.Platform, problem.ContestID, problem.Index)

	// Create directory
	if err := os.MkdirAll(problemDir, 0755); err != nil {
		return fmt.Errorf("failed to create problem dir: %w", err)
	}

	// Save problem.yaml
	problemPath := filepath.Join(problemDir, "problem.yaml")
	data, err := yaml.Marshal(problem)
	if err != nil {
		return fmt.Errorf("failed to marshal problem: %w", err)
	}
	if err := os.WriteFile(problemPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write problem: %w", err)
	}

	// Create tests directory and save samples
	testsDir := filepath.Join(problemDir, "tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		return fmt.Errorf("failed to create tests dir: %w", err)
	}

	for _, sample := range problem.Samples {
		inputPath := filepath.Join(testsDir, fmt.Sprintf("sample_%d.in", sample.Index))
		outputPath := filepath.Join(testsDir, fmt.Sprintf("sample_%d.out", sample.Index))

		if err := os.WriteFile(inputPath, []byte(sample.Input), 0644); err != nil {
			return fmt.Errorf("failed to write sample input: %w", err)
		}
		if err := os.WriteFile(outputPath, []byte(sample.Output), 0644); err != nil {
			return fmt.Errorf("failed to write sample output: %w", err)
		}
	}

	// Create solutions directory
	solutionsDir := filepath.Join(problemDir, "solutions")
	if err := os.MkdirAll(solutionsDir, 0755); err != nil {
		return fmt.Errorf("failed to create solutions dir: %w", err)
	}

	return nil
}

// LoadProblem loads a problem from the workspace
func (w *Workspace) LoadProblem(platform string, contestID int, index string) (*v1.Problem, error) {
	problemDir := w.ProblemPath(platform, contestID, index)
	problemPath := filepath.Join(problemDir, "problem.yaml")

	data, err := os.ReadFile(problemPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read problem: %w", err)
	}

	var problem v1.Problem
	if err := yaml.Unmarshal(data, &problem); err != nil {
		return nil, fmt.Errorf("failed to parse problem: %w", err)
	}

	return &problem, nil
}

// ProblemExists checks if a problem exists in the workspace
func (w *Workspace) ProblemExists(platform string, contestID int, index string) bool {
	problemPath := filepath.Join(w.ProblemPath(platform, contestID, index), "problem.yaml")
	_, err := os.Stat(problemPath)
	return err == nil
}

// ListProblems lists all problems in the workspace
func (w *Workspace) ListProblems() ([]*v1.Problem, error) {
	var problems []*v1.Problem

	problemsRoot := w.ProblemsPath()

	err := filepath.Walk(problemsRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.Name() == "problem.yaml" && !info.IsDir() {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil // Skip read errors
			}

			var problem v1.Problem
			if err := yaml.Unmarshal(data, &problem); err != nil {
				return nil // Skip parse errors
			}

			problems = append(problems, &problem)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return problems, nil
}

// SaveStatement saves the problem statement as markdown
func (w *Workspace) SaveStatement(problem *v1.Problem, statement string) error {
	problemDir := w.ProblemPath(problem.Platform, problem.ContestID, problem.Index)
	statementPath := filepath.Join(problemDir, "statement.md")

	// Format statement
	md := formatStatement(problem, statement)

	if err := os.WriteFile(statementPath, []byte(md), 0644); err != nil {
		return fmt.Errorf("failed to write statement: %w", err)
	}

	return nil
}

// LoadStatement loads the problem statement
func (w *Workspace) LoadStatement(platform string, contestID int, index string) (string, error) {
	problemDir := w.ProblemPath(platform, contestID, index)
	statementPath := filepath.Join(problemDir, "statement.md")

	data, err := os.ReadFile(statementPath)
	if err != nil {
		return "", fmt.Errorf("failed to read statement: %w", err)
	}

	return string(data), nil
}

// SaveNotes saves user notes for a problem
func (w *Workspace) SaveNotes(platform string, contestID int, index string, notes *v1.UserNotes) error {
	problem, err := w.LoadProblem(platform, contestID, index)
	if err != nil {
		return err
	}

	problem.Notes = *notes
	return w.SaveProblem(problem)
}

// UpdatePractice updates practice data for a problem
func (w *Workspace) UpdatePractice(platform string, contestID int, index string, practice *v1.PracticeData) error {
	problem, err := w.LoadProblem(platform, contestID, index)
	if err != nil {
		return err
	}

	problem.Practice = *practice
	return w.SaveProblem(problem)
}

func formatStatement(problem *v1.Problem, statement string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s. %s\n\n", problem.Index, problem.Name))
	sb.WriteString(fmt.Sprintf("**Rating:** %d | ", problem.Metadata.Rating))
	sb.WriteString(fmt.Sprintf("**Time:** %s | ", problem.Limits.TimeLimit))
	sb.WriteString(fmt.Sprintf("**Memory:** %s\n\n", problem.Limits.MemoryLimit))

	if len(problem.Metadata.Tags) > 0 {
		sb.WriteString("**Tags:** ")
		sb.WriteString(strings.Join(problem.Metadata.Tags, ", "))
		sb.WriteString("\n\n")
	}

	sb.WriteString("---\n\n")
	sb.WriteString("## Statement\n\n")
	sb.WriteString(statement)
	sb.WriteString("\n\n")

	if len(problem.Samples) > 0 {
		sb.WriteString("## Examples\n\n")
		for _, sample := range problem.Samples {
			sb.WriteString(fmt.Sprintf("### Example %d\n\n", sample.Index))
			sb.WriteString("**Input:**\n```\n")
			sb.WriteString(sample.Input)
			sb.WriteString("\n```\n\n")
			sb.WriteString("**Output:**\n```\n")
			sb.WriteString(sample.Output)
			sb.WriteString("\n```\n\n")
		}
	}

	sb.WriteString("---\n\n")
	sb.WriteString(fmt.Sprintf("[Open on Codeforces](%s)\n", problem.URL))

	return sb.String()
}
