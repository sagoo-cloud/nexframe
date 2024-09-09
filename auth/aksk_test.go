package auth

import (
	"testing"
)

func TestVerifySignature(t *testing.T) {
	type args struct {
		ak      string
		sk      string
		timeStr string
		sign    string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				ak:      "iotak",
				sk:      "iotsk20200907",
				timeStr: "1703834918",
				sign:    "90cfbaabe282e97228115f3684be4e6b626418cf2de18254d4fa0308a08d2044",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VerifySignature(tt.args.ak, tt.args.sk, tt.args.timeStr, tt.args.sign); got != tt.want {
				t.Errorf("VerifySignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateSignature(t *testing.T) {
	type args struct {
		message string
		secret  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				message: "ak=iotak&time=1703834918",
				secret:  "iotsk20200907",
			},
			want: "90cfbaabe282e97228115f3684be4e6b626418cf2de18254d4fa0308a08d2044",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateSignature(tt.args.message, tt.args.secret); got != tt.want {
				t.Errorf("GenerateSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}
