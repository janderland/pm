package state

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

type Task struct {
	Title string
	Story string
}

// Normalize cleans up the title & story of the
// task. Lead & trailing whitespace is removed
// and any substring of adjacent whitespace is
// replaced by a single space. The title is also
// changed to all caps. After these transformations,
// if either the title or story is an empty string,
// an error is returned.
func (t *Task) Normalize() error {
	title := strings.ToUpper(trim(t.Title))
	story := trim(t.Story)
	if title == "" {
		return errors.New("title is empty")
	}
	if story == "" {
		return errors.New("story is empty")
	}
	t.Title = title
	t.Story = story
	return nil
}

// Edit encodes the task into a temporary file and open the file in
// a text editor. If the editor exits with 0 exit code, then the
// temporary file is decode back into the task. If the editor fails,
// the task remains unedited. The editor is run in the shell specified
// with the environment variable $SHELL (defaulting to "sh"). The
// editor is specified with the environment variable $EDITOR (defaulting
// to "vim").
func (t *Task) Edit() error {
	namePattern := "*_" + strings.ReplaceAll(t.Title, " ", "_")
	file, err := ioutil.TempFile("", namePattern)
	if err != nil {
		return errors.Wrap(err, "failed to open temp file")
	}

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	err = enc.Encode(t)
	if err != nil {
		return errors.Wrap(err, "failed to encode task")
	}
	if err = file.Close(); err != nil {
		return errors.Wrap(err, "failed to close file")
	}

	shell := strings.TrimSpace(os.Getenv("SHELL"))
	if len(shell) == 0 {
		shell = "sh"
	}
	editor := strings.TrimSpace(os.Getenv("EDITOR"))
	if len(editor) == 0 {
		editor = "vim"
	}
	cmd := exec.Command(shell, "-c", fmt.Sprintf("%s %s", editor, file.Name()))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err = cmd.Run(); err != nil {
		return errors.Wrap(err, "failed to execute editor")
	}

	file, err = os.Open(file.Name())
	if err != nil {
		return errors.Wrap(err, "failed to open file")
	}
	dec := json.NewDecoder(file)
	if err = dec.Decode(t); err != nil {
		return errors.Wrap(err, "failed to decode file")
	}

	return nil
}

// trim transforms any adjacent whitespace into a single
// space and removes any leading or trailing whitespace.
func trim(str string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(str), " ")
}
