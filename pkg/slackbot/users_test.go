package slackbot

import (
	"path"
	"strings"
	"testing"
)

func TestSlackUserResolver_getSlackEmailFromMapping(t *testing.T) {
	testData := path.Join("test_data", "users")

	userMappingExist := make(map[string]string)
	userMappingExist["wine@yummy.com"] = "grapes@yummy.com"
	userMappingExist["beer@yummy.com"] = "margarita@yummy.com"

	type fields struct {
		UserMappings map[string]string
	}
	type args struct {
		gitUserEmail string
		fileLocation string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "find_from_existing_map",
			fields:  fields{userMappingExist},
			args:    args{gitUserEmail: "wine@yummy.com", fileLocation: "/foo"},
			want:    "grapes@yummy.com",
			wantErr: false,
		},
		{
			name:    "nil_user_mappings",
			fields:  fields{nil},
			args:    args{gitUserEmail: "wine@yummy.com", fileLocation: "/foo"},
			want:    "",
			wantErr: true,
			errMsg:  "failed to read file",
		},
		{
			name:    "nil_user_mappings_but_file_exists",
			fields:  fields{nil},
			args:    args{gitUserEmail: "wine@yummy.com", fileLocation: path.Join(testData, "user_mapping_file.txt")},
			want:    "grapes@yummy.com",
			wantErr: false,
		},
		{
			name:    "duplicate_git_user",
			fields:  fields{nil},
			args:    args{gitUserEmail: "wine@yummy.com", fileLocation: path.Join(testData, "user_mapping_file_duplicate.txt")},
			want:    "",
			wantErr: true,
			errMsg:  "duplicate mapping found for git user email",
		},
		{
			name:    "no_user_mappings",
			fields:  fields{nil},
			args:    args{gitUserEmail: "does_not_exist@yummy.com", fileLocation: path.Join(testData, "user_mapping_file.txt")},
			want:    "",
			wantErr: true,
			errMsg:  "no slack email found for git user email",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &SlackUserResolver{
				UserMappings: tt.fields.UserMappings,
			}
			got, err := r.getSlackEmailFromMapping(tt.args.gitUserEmail, tt.args.fileLocation)
			if err != nil {
				t.Logf("err %s", err.Error())
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("getSlackEmailFromMapping() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("getSlackEmailFromMapping() errMsg does not match got = %s, want %s", err.Error(), tt.errMsg)
				return
			}
			if got != tt.want {
				t.Errorf("getSlackEmailFromMapping() got = %v, want %v", got, tt.want)
			}
		})
	}
}
