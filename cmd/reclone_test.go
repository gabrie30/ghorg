package cmd

import (
	"reflect"
	"testing"
)

func Test_splitCommandArgs(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want []string
	}{
		{
			name: "simple command without quotes",
			cmd:  "ghorg clone my-org --scm=gitlab",
			want: []string{"ghorg", "clone", "my-org", "--scm=gitlab"},
		},
		{
			name: "double-quoted flag value with spaces",
			cmd:  `ghorg clone my-org --match-regex "(foo|bar)"`,
			want: []string{"ghorg", "clone", "my-org", "--match-regex", "(foo|bar)"},
		},
		{
			name: "double-quoted flag value with equals",
			cmd:  `ghorg clone my-org --match-regex="(foo|bar)"`,
			want: []string{"ghorg", "clone", "my-org", "--match-regex=(foo|bar)"},
		},
		{
			name: "single-quoted flag value",
			cmd:  `ghorg clone my-org --match-regex '(foo|bar)'`,
			want: []string{"ghorg", "clone", "my-org", "--match-regex", "(foo|bar)"},
		},
		{
			name: "gitlab-group-exclude-match-regex with double quotes (reported bug)",
			cmd:  `ghorg clone group1/group2 --gitlab-group-exclude-match-regex "(subgroup1|subgroup2|subgroup3|helm-charts)"`,
			want: []string{"ghorg", "clone", "group1/group2", "--gitlab-group-exclude-match-regex", "(subgroup1|subgroup2|subgroup3|helm-charts)"},
		},
		{
			name: "gitlab-group-match-regex with double quotes",
			cmd:  `ghorg clone group1/group2 --gitlab-group-match-regex "(subgroup1|subgroup2)"`,
			want: []string{"ghorg", "clone", "group1/group2", "--gitlab-group-match-regex", "(subgroup1|subgroup2)"},
		},
		{
			name: "multiple quoted arguments",
			cmd:  `ghorg clone my-org --match-regex "(foo|bar)" --exclude-match-regex "(baz|qux)"`,
			want: []string{"ghorg", "clone", "my-org", "--match-regex", "(foo|bar)", "--exclude-match-regex", "(baz|qux)"},
		},
		{
			name: "quoted value containing spaces",
			cmd:  `ghorg clone my-org --output-dir "my output dir"`,
			want: []string{"ghorg", "clone", "my-org", "--output-dir", "my output dir"},
		},
		{
			name: "no quotes at all",
			cmd:  "ghorg clone my-org --token=abc123 --scm=github",
			want: []string{"ghorg", "clone", "my-org", "--token=abc123", "--scm=github"},
		},
		{
			name: "complex regex with backslashes and anchors",
			cmd:  `ghorg clone my-group --gitlab-group-exclude-match-regex ".*\/subgroup-a($|\/.*$)"`,
			want: []string{"ghorg", "clone", "my-group", "--gitlab-group-exclude-match-regex", `.*\/subgroup-a($|\/.*$)`},
		},
		{
			name: "mixed quoted and unquoted flags",
			cmd:  `ghorg clone my-org --scm=gitlab --base-url=https://gitlab.example.com --token=secret --gitlab-group-match-regex "(team-a|team-b)" --output-dir=my-repos`,
			want: []string{"ghorg", "clone", "my-org", "--scm=gitlab", "--base-url=https://gitlab.example.com", "--token=secret", "--gitlab-group-match-regex", "(team-a|team-b)", "--output-dir=my-repos"},
		},
		{
			name: "multiple spaces between arguments",
			cmd:  "ghorg  clone  my-org",
			want: []string{"ghorg", "clone", "my-org"},
		},
		{
			name: "single quotes inside double quotes are preserved",
			cmd:  `ghorg clone my-org --match-regex "it's-a-test"`,
			want: []string{"ghorg", "clone", "my-org", "--match-regex", "it's-a-test"},
		},
		{
			name: "double quotes inside single quotes are preserved",
			cmd:  `ghorg clone my-org --match-regex 'say "hello"'`,
			want: []string{"ghorg", "clone", "my-org", "--match-regex", `say "hello"`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitCommandArgs(tt.cmd)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitCommandArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sanitizeCmd(t *testing.T) {
	type args struct {
		cmd string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "shorthand with space",
			args: args{cmd: "ghorg clone foo -t bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2"},
			want: "ghorg clone foo -t XXXXXXX ",
		},
		{
			name: "shorthand with equals",
			args: args{cmd: "ghorg clone foo -t=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2"},
			want: "ghorg clone foo -t=XXXXXXX ",
		},
		{
			name: "longhand with space",
			args: args{cmd: "ghorg clone foo --token bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2"},
			want: "ghorg clone foo --token XXXXXXX ",
		},
		{
			name: "longhand with equals",
			args: args{cmd: "ghorg clone foo --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2"},
			want: "ghorg clone foo --token=XXXXXXX ",
		},
		{
			name: "shorthand with equals does not pick up other flags with t",
			args: args{cmd: "ghorg clone foo -t=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2 --topics=foo,bar"},
			want: "ghorg clone foo -t=XXXXXXX --topics=foo,bar",
		},
		{
			name: "shorthand with space does not pick up other flags with t",
			args: args{cmd: "ghorg clone foo -t bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2 --topics=foo,bar"},
			want: "ghorg clone foo -t XXXXXXX --topics=foo,bar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeCmd(tt.args.cmd); got != tt.want {
				t.Errorf("sanitizeCmd() = %v, want %v", got, tt.want)
			}
		})
	}
}
