package cmd

import "testing"

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
