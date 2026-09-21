//go:build linux

package agent

import "testing"

func TestLinuxOneAPIWrappedCommand(
	t *testing.T,
) {
	command,
		args,
		err :=
		linuxOneAPIWrappedCommand(
			"/usr/bin/time",
			[]string{
				"-v",
				"/tmp/llama-bench",
				"-m",
				"/tmp/model with spaces.gguf",
				"--example",
				"value;not-shell",
			},
		)

	if err != nil {
		t.Fatal(err)
	}

	if command !=
		m9LinuxBashPath {

		t.Fatalf(
			"command = %q",
			command,
		)
	}

	want :=
		[]string{
			"-c",
			m9LinuxOneAPICommandScript,
			"meshalot-oneapi",
			m9LinuxOneAPISetvarsPath,
			"/usr/bin/time",
			"-v",
			"/tmp/llama-bench",
			"-m",
			"/tmp/model with spaces.gguf",
			"--example",
			"value;not-shell",
		}

	if len(args) != len(want) {
		t.Fatalf(
			"argument count = %d, want %d",
			len(args),
			len(want),
		)
	}

	for i := range want {
		if args[i] != want[i] {
			t.Fatalf(
				"argument %d = %q, want %q",
				i,
				args[i],
				want[i],
			)
		}
	}
}

func TestLinuxOneAPIWrappedCommandRejectsEmptyCommand(
	t *testing.T,
) {
	if _, _, err :=
		linuxOneAPIWrappedCommand(
			" ",
			nil,
		); err == nil {

		t.Fatal(
			"empty Linux command was accepted",
		)
	}
}
