package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/ofirmad/task-manager/models"
	"github.com/ofirmad/task-manager/services"
	"github.com/ofirmad/task-manager/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestRequester(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteConfig, reporterConfig := GinkgoConfiguration()
	reporterConfig.JUnitReport = "tests.xml"
	RunSpecs(t, "Handle Tasks", suiteConfig, reporterConfig)
}

const (
	tasksPath = "/tasks"
)

var _ = Describe("Handle Tasks Tests", func() {
	var task models.Task

	BeforeEach(func() {
		task = models.Task{
			Title:       "New Task",
			Description: "Task Description",
			Status:      "Pending",
		}
	})

	Describe("GET /tasks", func() {
		BeforeEach(func() {
			// Reset the in-memory database before each test
			models.DB.Tasks = make(map[int]*models.Task)
			models.DB.NextID = 1
		})

		It("should successfully return empty tasks list when there are no tasks", func() {
			response := performRequest(http.MethodGet, tasksPath, nil)
			Expect(response.Code).To(Equal(http.StatusOK))

			var responseBody []map[string]interface{}
			Expect(json.Unmarshal(response.Body.Bytes(), &responseBody)).To(Succeed())
			Expect(responseBody).To(BeEmpty())
			Expect(responseBody).To(HaveLen(0))
		})

		It("should successfully a task when there is one task", func() {
			services.CreateTask(task)

			response := performRequest(http.MethodGet, tasksPath, nil)
			Expect(response.Code).To(Equal(http.StatusOK))

			var responseBody []map[string]interface{}
			Expect(json.Unmarshal(response.Body.Bytes(), &responseBody)).To(Succeed())
			Expect(responseBody).To(HaveLen(1))
			testutils.ValidateResponse(task, responseBody[0])
		})

		It("should successfully return multiple tasks when there are multiple tasks", func() {
			task1 := models.Task{
				Title:       task.Title + " 1",
				Description: task.Description + " 1",
				Status:      task.Status,
			}
			services.CreateTask(task1)

			task2 := models.Task{
				Title:       task.Title + " 2",
				Description: task.Description + " 2",
				Status:      "Completed",
			}
			services.CreateTask(task2)

			response := performRequest(http.MethodGet, tasksPath, nil)
			Expect(response.Code).To(Equal(http.StatusOK))

			var responseBody []map[string]interface{}
			Expect(json.Unmarshal(response.Body.Bytes(), &responseBody)).To(Succeed())
			Expect(responseBody).To(HaveLen(2))
			testutils.ValidateResponse(task1, responseBody[0])
			testutils.ValidateResponse(task2, responseBody[1])
		})
	})

	Describe("POST /tasks", func() {
		It("should successfully create a new task", func() {
			response := performRequest(http.MethodPost, tasksPath, task)
			Expect(response.Code).To(Equal(http.StatusCreated))

			var responseBody map[string]interface{}
			Expect(json.Unmarshal(response.Body.Bytes(), &responseBody)).To(Succeed())

			testutils.ValidateResponse(task, responseBody)
		})

		It("should fail to create a new task with invalid request payload - title is missing", func() {
			task := models.Task{
				Description: "Task Description",
				Status:      "Pending",
			}

			response := performRequest(http.MethodPost, tasksPath, task)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring(titleRequired))
		})

		It("should fail to create a new task with invalid request payload - description is missing", func() {
			task := models.Task{
				Title:  "New Task",
				Status: "Pending",
			}

			response := performRequest(http.MethodPost, tasksPath, task)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring(descriptionRequired))
		})

		It("should fail to create a new task with invalid request payload - status is missing", func() {
			task := models.Task{
				Title:       "New Task",
				Description: "Task Description",
			}

			response := performRequest(http.MethodPost, tasksPath, task)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring(statusRequired))
		})

		It("should fail to create a new task with invalid request payload - invalid status", func() {
			task := models.Task{
				Title:       "New Task",
				Description: "Task Description",
				Status:      "Invalid",
			}

			response := performRequest(http.MethodPost, tasksPath, task)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring(invalidStatus))
		})
	})
})

var _ = Describe("Handle Task By ID Tests", func() {
	var task models.Task

	BeforeEach(func() {
		// Reset the in-memory database before each test
		models.DB.Tasks = make(map[int]*models.Task)
		models.DB.NextID = 1

		task = models.Task{
			Title:       "New Task",
			Description: "Task Description",
			Status:      "Pending",
		}
	})

	Describe("GET /tasks/{id}", func() {
		It("should successfully return a task by ID", func() {
			newTask := services.CreateTask(task)

			response := performRequest(http.MethodGet, tasksPath+"/"+strconv.Itoa(newTask.ID), nil)
			Expect(response.Code).To(Equal(http.StatusOK))

			var responseBody map[string]interface{}
			Expect(json.Unmarshal(response.Body.Bytes(), &responseBody)).To(Succeed())
			testutils.ValidateResponse(task, responseBody)
		})

		It("should fail to return a task by ID when the task does not exist", func() {
			response := performRequest(http.MethodGet, tasksPath+"/1", nil)
			Expect(response.Code).To(Equal(http.StatusNotFound))
			Expect(response.Body.String()).To(ContainSubstring(services.TaskNotFound))
		})
	})

	Describe("PUT /tasks/{id}", func() {
		It("should successfully update a task by ID", func() {
			// add a task
			newTask := services.CreateTask(task)

			// update the task
			updatedTask := models.Task{
				Title:       "Updated Task",
				Description: "Updated Task Description",
				Status:      "Completed",
			}

			response := performRequest(http.MethodPut, tasksPath+"/"+strconv.Itoa(newTask.ID), updatedTask)
			Expect(response.Code).To(Equal(http.StatusOK))

			var responseBody map[string]interface{}
			Expect(json.Unmarshal(response.Body.Bytes(), &responseBody)).To(Succeed())

			// validate id and created_at did not change
			testutils.ValidateIDAndCreatedAt(newTask, responseBody)

			testutils.ValidateResponse(updatedTask, responseBody)
		})

		It("should fail to update a task by ID dut to invalid request payload - title is missing", func() {
			newTask := services.CreateTask(task)

			updatedTask := models.Task{
				Description: "Updated Task Description",
				Status:      "Completed",
			}

			response := performRequest(http.MethodPut, tasksPath+"/"+strconv.Itoa(newTask.ID), updatedTask)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring(titleRequired))

			// get the task to verify it was not updated
			response = performRequest(http.MethodGet, tasksPath+"/"+strconv.Itoa(newTask.ID), nil)
			Expect(response.Code).To(Equal(http.StatusOK))

			var responseBody map[string]interface{}
			Expect(json.Unmarshal(response.Body.Bytes(), &responseBody)).To(Succeed())

			testutils.ValidateResponse(newTask, responseBody)
			testutils.ValidateIDAndCreatedAt(newTask, responseBody)
		})

		It("should fail to update a task by ID dut to invalid request payload - description is missing", func() {
			newTask := services.CreateTask(task)

			updatedTask := models.Task{
				Title:  "Updated Task",
				Status: "Completed",
			}

			response := performRequest(http.MethodPut, tasksPath+"/"+strconv.Itoa(newTask.ID), updatedTask)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring(descriptionRequired))

			// get the task to verify it was not updated
			response = performRequest(http.MethodGet, tasksPath+"/"+strconv.Itoa(newTask.ID), nil)
			Expect(response.Code).To(Equal(http.StatusOK))

			var responseBody map[string]interface{}
			Expect(json.Unmarshal(response.Body.Bytes(), &responseBody)).To(Succeed())

			testutils.ValidateResponse(newTask, responseBody)
			testutils.ValidateIDAndCreatedAt(newTask, responseBody)
		})

		It("should fail to update a task by ID dut to invalid request payload - status is missing", func() {
			newTask := services.CreateTask(task)

			updatedTask := models.Task{
				Title:       "Updated Task",
				Description: "Updated Task Description",
			}

			response := performRequest(http.MethodPut, tasksPath+"/"+strconv.Itoa(newTask.ID), updatedTask)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring(statusRequired))

			// get the task to verify it was not updated
			response = performRequest(http.MethodGet, tasksPath+"/"+strconv.Itoa(newTask.ID), nil)
			Expect(response.Code).To(Equal(http.StatusOK))

			var responseBody map[string]interface{}
			Expect(json.Unmarshal(response.Body.Bytes(), &responseBody)).To(Succeed())

			testutils.ValidateResponse(newTask, responseBody)
			testutils.ValidateIDAndCreatedAt(newTask, responseBody)
		})

		It("should fail to update a task by ID dut to invalid request payload - invalid status", func() {
			newTask := services.CreateTask(task)

			updatedTask := models.Task{
				Title:       "Updated Task",
				Description: "Updated Task Description",
				Status:      "Invalid",
			}

			response := performRequest(http.MethodPut, tasksPath+"/"+strconv.Itoa(newTask.ID), updatedTask)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring(invalidStatus))

			// get the task to verify it was not updated
			response = performRequest(http.MethodGet, tasksPath+"/"+strconv.Itoa(newTask.ID), nil)
			Expect(response.Code).To(Equal(http.StatusOK))

			var responseBody map[string]interface{}
			Expect(json.Unmarshal(response.Body.Bytes(), &responseBody)).To(Succeed())

			testutils.ValidateResponse(newTask, responseBody)
			testutils.ValidateIDAndCreatedAt(newTask, responseBody)
		})

		It("should fail to update a task by ID when the task does not exist", func() {
			response := performRequest(http.MethodPut, tasksPath+"/1", task)
			Expect(response.Code).To(Equal(http.StatusNotFound))
			Expect(response.Body.String()).To(ContainSubstring(services.TaskNotFound))
		})
	})

	Describe("DELETE /tasks/{id}", func() {
		It("should successfully delete a task by ID", func() {
			newTask := services.CreateTask(task)

			response := performRequest(http.MethodDelete, tasksPath+"/"+strconv.Itoa(newTask.ID), nil)
			Expect(response.Code).To(Equal(http.StatusNoContent))

			var responseBody map[string]interface{}
			Expect(json.Unmarshal(response.Body.Bytes(), &responseBody)).NotTo(Succeed())
			Expect(responseBody).To(BeEmpty())

			// verify the task was deleted
			response = performRequest(http.MethodGet, tasksPath+"/"+strconv.Itoa(newTask.ID), nil)
			Expect(response.Code).To(Equal(http.StatusNotFound))
			Expect(response.Body.String()).To(ContainSubstring(services.TaskNotFound))
		})

		It("should fail to delete a task by ID when the task does not exist", func() {
			response := performRequest(http.MethodDelete, tasksPath+"/1", nil)
			Expect(response.Code).To(Equal(http.StatusNotFound))
			Expect(response.Body.String()).To(ContainSubstring(services.TaskNotFound))
		})
	})
})

func performRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var requestBody []byte
	if body != nil {
		requestBody, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Use the actual handlers from the app
	w := httptest.NewRecorder()
	if path == tasksPath {
		HandleTasks(w, req)
	} else {
		HandleTaskByID(w, req)
	}
	return w
}
