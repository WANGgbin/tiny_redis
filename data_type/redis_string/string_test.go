package redis_string

import "testing"

func TestStrCmp(t *testing.T) {
	type args struct {
		src *RedisString
		dst *RedisString
	}
	tests := []struct {
		name    string
		args    args
		want    int8
		wantErr bool
	}{
		{
			name: "dst is nil",
			args: args{
				src: &RedisString{
					Content: []byte("abc"),
				},
				dst: nil,
			},
			want: int8(-1),
			wantErr: true,
		},
		{
			name: "Content of dst is nil",
			args: args{
				src: &RedisString{
					Content: []byte("abc"),
				},
				dst: &RedisString{
					Content: nil,
				},
			},
			want: int8(1),
			wantErr: false,
		},
		{
			name: "src is bigger then src cause logger length",
			args: args{
				src: &RedisString{
					Content: []byte("abc"),
				},
				dst: &RedisString{
					Content: []byte("ab"),
				},
			},
			want: int8(1),
			wantErr: false,
		},
		{
			name: "src is bigger then src with same length",
			args: args{
				src: &RedisString{
					Content: []byte("Abc\x00"),
				},
				dst: &RedisString{
					Content: []byte("abc\x01"),
				},
			},
			want: int8(1),
			wantErr: false,
		},
		{
			name: "src equals to dst",
			args: args{
				src: &RedisString{
					Content: []byte("abc"),
				},
				dst: &RedisString{
					Content: []byte("abc"),
				},
			},
			want: int8(0),
			wantErr: false,
		},
		{
			name: "dst is bigger then src with same length",
			args: args{
				src: &RedisString{
					Content: []byte("abc"),
				},
				dst: &RedisString{
					Content: []byte("Abc"),
				},
			},
			want: int8(-1),
			wantErr: false,
		},
		{
			name: "dst is bigger then src cause logger length",
			args: args{
				src: &RedisString{
					Content: []byte("abc"),
				},
				dst: &RedisString{
					Content: []byte("abc\x01"),
				},
			},
			want: int8(-1),
			wantErr: false,
		},
		
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StrCmp(tt.args.src, tt.args.dst)
			if (err != nil) != tt.wantErr {
				t.Errorf("StrCmp(%s, %s) error = %v, wantErr %v", string(tt.args.src.Content), string(tt.args.dst.Content), err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StrCmp(%s, %s) = %v, want %v", string(tt.args.src.Content), string(tt.args.dst.Content), got, tt.want)
			}
		})
	}
}
