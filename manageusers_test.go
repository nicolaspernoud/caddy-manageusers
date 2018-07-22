package manageusers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestAPISequenceWithoutUserfile(t *testing.T) {
	userInfoFile := ""
	expectedFromGet := `{"admin":{"username":"admin","password":"","origin":"","role":"","name":"","surname":""},"user":{"username":"user","password":"","origin":"","role":"","name":"","surname":""}}`
	expectedFromGetAfterPost := `{"admin":{"username":"admin","password":"","origin":"","role":"","name":"","surname":""},"tester":{"username":"tester","password":"","origin":"","role":"","name":"","surname":""},"user":{"username":"user","password":"","origin":"","role":"","name":"","surname":""}}`
	doGET(t, userInfoFile, expectedFromGet)
	doBadPOST(t, userInfoFile)
	doPOST(t, userInfoFile)
	doGET(t, userInfoFile, expectedFromGetAfterPost)
	doBadDELETE(t, userInfoFile)
	doDELETE(t, userInfoFile)
	doGET(t, userInfoFile, expectedFromGet)
}

func TestAPISequenceWithUserInfoFile(t *testing.T) {
	userInfoFile := "./testdata/users.json"
	expectedFromGet := `{"admin":{"username":"admin","password":"","origin":"htpasswd","role":"admin","name":"Ad","surname":"MIN"},"user":{"username":"user","password":"","origin":"htpasswd","role":"user","name":"Us","surname":"ER"}}`
	expectedFromGetAfterPost := `{"admin":{"username":"admin","password":"","origin":"htpasswd","role":"admin","name":"Ad","surname":"MIN"},"tester":{"username":"tester","password":"","origin":"htpasswd","role":"tester","name":"Test","surname":"ER"},"user":{"username":"user","password":"","origin":"htpasswd","role":"user","name":"Us","surname":"ER"}}`
	doGET(t, userInfoFile, expectedFromGet)
	doBadPOST(t, userInfoFile)
	doPOST(t, userInfoFile)
	doGET(t, userInfoFile, expectedFromGetAfterPost)
	doBadDELETE(t, userInfoFile)
	doDELETE(t, userInfoFile)
	doGET(t, userInfoFile, expectedFromGet)
}

func doGET(t *testing.T, userInfoFile string, expected string) {
	// Create a response recorder to listen for the response
	rr := httptest.NewRecorder()
	handler := PwdHandler{
		Config: Config{
			Route:        "/manageusers",
			HtPasswdFile: "./testdata/test.htpasswd",
			UserInfoFile: userInfoFile,
		},
		Next: nil,
	}

	// Create a request to pass to handler
	req, err := http.NewRequest("GET", "/manageusers", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body
	got := strings.Trim(rr.Body.String(), " \n")
	if got != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			got, expected)
	}
}

func doPOST(t *testing.T, userInfoFile string) {

	// Create a response recorder to listen for the response
	rr := httptest.NewRecorder()
	handler := PwdHandler{
		Config: Config{
			Route:        "/manageusers",
			HtPasswdFile: "./testdata/test.htpasswd",
			UserInfoFile: userInfoFile,
		},
		Next: nil,
	}

	// Create a request to pass to handler
	u := User{Username: "tester", Password: "testerpwd", Origin: "htpasswd", Role: "tester", Name: "Test", Surname: "ER"}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(u)
	req, err := http.NewRequest("POST", "/manageusers", b)
	if err != nil {
		t.Fatal(err)
	}

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body
	expected := `User tester added or altered`
	got := strings.Trim(rr.Body.String(), " \n")
	if got != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			got, expected)
	}
}

func doBadPOST(t *testing.T, userInfoFile string) {

	// Create a response recorder to listen for the response
	rr := httptest.NewRecorder()
	handler := PwdHandler{
		Config: Config{
			Route:        "/manageusers",
			HtPasswdFile: "./testdata/test.htpasswd",
			UserInfoFile: userInfoFile,
		},
		Next: nil,
	}

	// Create a request to pass to handler
	var jsonStr = []byte(`{
		"username": "userwithemptypassword",
		"password": "",
		"role": "user",
		"name": "Bad",
		"surname": "USER"
	}`)
	req, err := http.NewRequest("POST", "/manageusers", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	// Check the response body
	expected := `Error adding or altering user userwithemptypassword`
	got := strings.Trim(rr.Body.String(), " \n")
	if got != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			got, expected)
	}
}

func doDELETE(t *testing.T, userInfoFile string) {

	// Create a response recorder to listen for the response
	rr := httptest.NewRecorder()
	handler := PwdHandler{
		Config: Config{
			Route:        "/manageusers",
			HtPasswdFile: "./testdata/test.htpasswd",
			UserInfoFile: userInfoFile,
		},
		Next: nil,
	}

	// Create a request to pass to handler
	req, err := http.NewRequest("DELETE", "/manageusers/tester", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body
	expected := `User tester deleted`
	got := strings.Trim(rr.Body.String(), " \n")
	if got != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			got, expected)
	}
}

func doBadDELETE(t *testing.T, userInfoFile string) {

	// Create a response recorder to listen for the response
	rr := httptest.NewRecorder()
	handler := PwdHandler{
		Config: Config{
			Route:        "/manageusers",
			HtPasswdFile: "./testdata/test.htpasswd",
			UserInfoFile: userInfoFile,
		},
		Next: nil,
	}

	// Create a request to pass to handler
	req, err := http.NewRequest("DELETE", "/manageusers/unexistinguser", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	// Check the response body
	expected := `Error deleting user unexistinguser`
	got := strings.Trim(rr.Body.String(), " \n")
	if got != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			got, expected)
	}
}

func TestParseUserInfoFile(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name      string
		args      args
		wantUsers []UserInfos
		wantErr   bool
	}{
		{
			"Get users from ok file",
			args{
				file: "./testdata/users.json",
			},
			[]UserInfos{
				{
					Sub:    "user",
					Origin: "htpasswd",
					Claims: Claims{
						Role:    "user",
						Name:    "Us",
						Surname: "ER",
					},
				},
				{
					Sub:    "admin",
					Origin: "htpasswd",
					Claims: Claims{
						Role:    "admin",
						Name:    "Ad",
						Surname: "MIN",
					},
				},
			},
			false,
		},
		{
			"Missing file",
			args{
				file: "./testdata/users.missing.json",
			},
			[]UserInfos{},
			true,
		},
		{
			"Badly formed file",
			args{
				file: "./testdata/users.bad.json",
			},
			[]UserInfos{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUsers, err := ParseUserInfoFile(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseUserInfoFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(gotUsers, tt.wantUsers) {
				t.Errorf("ParseUserInfoFile() = %v, want %v", gotUsers, tt.wantUsers)
			}
		})
	}
}

func Test_mergeUsers(t *testing.T) {
	emptyMap := make(map[string]User)
	type args struct {
		As *map[string]User
		Bs []UserInfos
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"Merge some users with users info",
			args{
				As: &map[string]User{
					"user": {
						Username: "user",
						Password: "",
						Role:     "",
						Name:     "",
						Surname:  "",
					},
					"admin": {
						Username: "admin",
						Password: "",
						Role:     "not",
						Name:     "to be",
						Surname:  "kept",
					},
				},
				Bs: []UserInfos{
					{
						Sub:    "user",
						Origin: "htpasswd",
						Claims: Claims{
							Role:    "user",
							Name:    "Us",
							Surname: "ER",
						},
					},
					{
						Sub:    "admin",
						Origin: "htpasswd",
						Claims: Claims{
							Role:    "admin",
							Name:    "Ad",
							Surname: "MIN",
						},
					},
				},
			},
			false,
		},
		{
			"User array is empty, should return error",
			args{
				As: &emptyMap,
				Bs: []UserInfos{
					{
						Sub:    "user",
						Origin: "htpasswd",
						Claims: Claims{
							Role:    "user",
							Name:    "Us",
							Surname: "ER",
						},
					},
					{
						Sub:    "admin",
						Origin: "htpasswd",
						Claims: Claims{
							Role:    "admin",
							Name:    "Ad",
							Surname: "MIN",
						},
					},
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := mergeUsers(tt.args.As, tt.args.Bs); (err != nil) != tt.wantErr {
				t.Errorf("mergeUsers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
