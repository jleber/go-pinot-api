package goPinotAPI_test

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	goPinotAPI "github.com/azaurus1/go-pinot-api"
	model "github.com/azaurus1/go-pinot-api/model"
	"github.com/stretchr/testify/assert"
)

const (
	RouteUsers                = "/users"
	RouteUser                 = "/users/test"
	RouteInstances            = "/instances"
	RouteInstance             = "/instances/Minion_172.19.0.2_9514"
	RouteClusterInfo          = "/cluster/info"
	RouteClusterConfigs       = "/cluster/configs"
	RouteClusterConfigsDelete = "/cluster/configs/allowParticipantAutoJoin"
)

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Basic YWRtaW46YWRtaW4K" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}

}

func handleGetUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `{"users": {"test_BROKER": {"username": "test","password": "$2a$10$3KYPvIy4fBM3CWfdKSc54u/BOn1rPlgb7u4P66s4upqAct30C/q6a","component": "BROKER","role": "ADMIN","usernameWithComponent": "test_BROKER"}}}`)
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	component := r.URL.Query().Get("component")

	if component == "BROKER" {
		fmt.Fprint(w, `{"test_BROKER": {"username": "test","password": "$2a$10$3KYPvIy4fBM3CWfdKSc54u/BOn1rPlgb7u4P66s4upqAct30C/q6a","component": "BROKER","role": "ADMIN","usernameWithComponent": "test_BROKER"}}`)
	}
}

func handlePutUser(w http.ResponseWriter, r *http.Request) {
	component := r.URL.Query().Get("component")

	if component == "BROKER" {
		fmt.Fprint(w, `{"status": "User config update for test_BROKER"}`)
	}

}

func handleDeleteUser(w http.ResponseWriter, r *http.Request) {

	component := r.URL.Query().Get("component")

	if component == "BROKER" {
		fmt.Fprint(w, `{"status": "User: test_BROKER has been successfully deleted"}`)
	} else if component == "" {
		http.Error(w, `{"code": 400,"error": "Name is null"}`, http.StatusBadRequest)
	}
}

func handlePostUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `{"status": "User testUser_BROKER has been successfully added!"}`)
}

func handleGetInstances(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `{"instances": ["Minion_172.19.0.2_9514","Server_172.19.0.7_8098","Broker_cdba1ba98e74_8099","Controller_8684b6757488_9000"]}`)
}

func handlePostInstances(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `{"status": "Added instance: Broker_localhost_1234"}`)
}

func handleGetInstance(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `{"instanceName": "Minion_172.19.0.2_9514","hostName": "172.19.0.2","enabled": true,"port": "9514","tags": ["minion_untagged"],"pools": null,"grpcPort": -1,"adminPort": -1,"queryServicePort": -1,"queryMailboxPort": -1,"systemResourceInfo": null}`)

}

func handleGetClusterInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `{"clusterName": "PinotCluster"}`)
}

func handleGetClusterConfigs(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `{"allowParticipantAutoJoin": "true","enable.case.insensitive": "true","pinot.broker.enable.query.limit.override": "false","default.hyperloglog.log2m": "8"}`)
}

func handlePostClusterConfigs(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `{"status": "Updated cluster config."}`)
}

func handleDeleteClusterConfigs(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `{"status": "Deleted cluster config: allowParticipantAutoJoin"}`)
}

func createMockControllerServer() *httptest.Server {

	mux := http.NewServeMux()

	mux.HandleFunc(RouteUsers, authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handleGetUsers(w, r)
		case "POST":
			handlePostUsers(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc(RouteUser, authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handleGetUser(w, r)
		case "PUT":
			handlePutUser(w, r)
		case "DELETE":
			handleDeleteUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc(RouteInstances, authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handleGetInstances(w, r)
		case "POST":
			handlePostInstances(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc(RouteInstance, authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handleGetInstance(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc(RouteClusterInfo, authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handleGetClusterInfo(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc(RouteClusterConfigs, authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handleGetClusterConfigs(w, r)
		case "POST":
			handlePostClusterConfigs(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc(RouteClusterConfigsDelete, authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "DELETE":
			handleDeleteClusterConfigs(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	return httptest.NewServer(mux)

}

func createPinotClient(server *httptest.Server) *goPinotAPI.PinotAPIClient {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	return goPinotAPI.NewPinotAPIClient(
		goPinotAPI.ControllerUrl(server.URL),
		goPinotAPI.AuthToken("YWRtaW46YWRtaW4K"),
		goPinotAPI.Logger(logger),
	)
}

// TestFetchData
// func TestFetchData(t *testing.T) {

// }

// test FetchData unauthorized

// Test GetUsers
func TestGetUsers(t *testing.T) {

	server := createMockControllerServer()
	client := createPinotClient(server)

	res, err := client.GetUsers()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// expect 1 user in the response username test
	assert.Equal(t, len(res.Users), 1, "Expected 1 user in the response")
	assert.Equal(t, res.Users["test_BROKER"].Username, "test", "Expected username to be test")

}

// Test GetUser
func TestGetUser(t *testing.T) {

	server := createMockControllerServer()
	client := createPinotClient(server)

	res, err := client.GetUser("test", "BROKER")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// expect username test
	// expect component BROKER
	assert.Equal(t, res.Username, "test", "Expected username to be test")
	assert.Equal(t, res.Component, "BROKER", "Expected component to be BROKER")

}

// Test CreateUser
func TestCreateUser(t *testing.T) {

	server := createMockControllerServer()
	client := createPinotClient(server)

	user := model.User{
		Username:  "testUser",
		Password:  "test",
		Component: "BROKER",
		Role:      "ADMIN",
	}

	userBytes, err := json.Marshal(user)
	if err != nil {
		t.Errorf("Couldn't marshal user: %v", err)
	}

	res, err := client.CreateUser(userBytes)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// expect User testUser_BROKER has been successfully added!
	assert.Equal(t, res.Status, "User testUser_BROKER has been successfully added!", "Expected response to be User testUser_BROKER has been successfully added!")
}

// Test UpdateUser
func TestUpdateUser(t *testing.T) {
	server := createMockControllerServer()
	client := createPinotClient(server)

	user := model.User{
		Username:  "test",
		Password:  "test",
		Component: "BROKER",
		Role:      "USER",
	}

	userBytes, err := json.Marshal(user)
	if err != nil {
		t.Errorf("Couldn't marshal user: %v", err)
	}

	res, err := client.UpdateUser(user.Username, user.Component, true, userBytes)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// expect User test_BROKER has been successfully updated!
	assert.Equal(t, res.Status, "User config update for test_BROKER", "Expected response to be User config update for test_BROKER")
}

// Test DeleteUser
func TestDeleteUser(t *testing.T) {
	server := createMockControllerServer()
	client := createPinotClient(server)

	res, err := client.DeleteUser("test", "BROKER")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// expect User test_BROKER has been successfully deleted!
	assert.Equal(t, res.Status, "User: test_BROKER has been successfully deleted", "Expected response to be User: test_BROKER has been successfully deleted!")

}

func TestDeleteUserNoComponent(t *testing.T) {
	server := createMockControllerServer()
	client := createPinotClient(server)

	_, err := client.DeleteUser("test", "")
	if err != nil {
		assert.Equal(t, err.Error(), "client: request failed: status 400\n{\"code\": 400,\"error\": \"Name is null\"}\n", "Expected error to be Name is null")
	}

}

func TestGetInstances(t *testing.T) {
	server := createMockControllerServer()
	client := createPinotClient(server)

	res, err := client.GetInstances()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	assert.Equal(t, len(res.Instances), 4, "Expected 4 instances in the response")
}

func TestGetInstance(t *testing.T) {
	server := createMockControllerServer()
	client := createPinotClient(server)

	res, err := client.GetInstance("Minion_172.19.0.2_9514")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	assert.Equal(t, res.InstanceName, "Minion_172.19.0.2_9514", "Expected instance name to be Minion_172.19.0.2_9514")
}

func TestCreateInstance(t *testing.T) {
	server := createMockControllerServer()
	client := createPinotClient(server)

	instance := model.Instance{
		Host: "localhost",
		Port: 1234,
		Type: "BROKER",
	}

	instanceBytes, err := json.Marshal(instance)
	if err != nil {
		t.Errorf("Couldn't marshal instance: %v", err)
	}

	res, err := client.CreateInstance(instanceBytes)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	assert.Equal(t, res.Status, "Added instance: Broker_localhost_1234", "Expected response to be Added instance: Broker_localhost_1234")

}

// DeleteInstance

func TestGetClusterInfo(t *testing.T) {
	server := createMockControllerServer()
	client := createPinotClient(server)

	res, err := client.GetClusterInfo()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	assert.Equal(t, res.ClusterName, "PinotCluster", "Expected cluster name to be PinotCluster")
}

func TestGetClusterConfigs(t *testing.T) {
	server := createMockControllerServer()
	client := createPinotClient(server)

	res, err := client.GetClusterConfigs()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	assert.Equal(t, res.AllowParticipantAutoJoin, "true", "Expected allowParticipantAutoJoin to be true")
}

func TestUpdateClusterConfig(t *testing.T) {
	server := createMockControllerServer()
	client := createPinotClient(server)

	config := model.ClusterConfig{
		AllowParticipantAutoJoin: "false",
	}

	configBytes, err := json.Marshal(config)
	if err != nil {
		t.Errorf("Couldn't marshal config: %v", err)
	}

	res, err := client.UpdateClusterConfigs(configBytes)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	assert.Equal(t, res.Status, "Updated cluster config.", "Expected response to be Updated cluster config.")

}

func TestDeleteClusterConfig(t *testing.T) {
	server := createMockControllerServer()
	client := createPinotClient(server)

	res, err := client.DeleteClusterConfig("allowParticipantAutoJoin")
	if err != nil {
		assert.Equal(t, err.Error(), "client: request failed: status 404\n404 page not found\n", "Expected error to be 404 page not found")
	}

	assert.Equal(t, res.Status, "Deleted cluster config: allowParticipantAutoJoin", "Expected response to be Deleted cluster config: allowParticipantAutoJoin")
}
