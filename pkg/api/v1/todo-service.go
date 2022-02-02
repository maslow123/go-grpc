package v1

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// apiVersion is version of API is provided by server
	apiVersion = "v1"
)

// todoServiceServer is implementation of v1.TodoServiceServer proto interface
type todoServiceServer struct {
	db *sql.DB
}

// NewTodoServiceServer creates Todo Service
func NewTodoServiceServer(db *sql.DB) TodoServiceServer {
	return &todoServiceServer{db: db}
}

// CheckAPI cheks if the API version requested by client is supported by server
func (s *todoServiceServer) checkAPI(api string) error {
	// API version is "" means use current version of the service
	if len(api) > 0 {
		if apiVersion != api {
			return status.Errorf(codes.Unimplemented,
				"Unsupported API version: service implements API version '%s', but asked for '%s'", apiVersion, api,
			)
		}
	}

	return nil
}

// connect returns SQL database connection from the pool
func (s *todoServiceServer) connect(ctx context.Context) (*sql.Conn, error) {
	c, err := s.db.Conn(ctx)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to connect to database -> "+err.Error())
	}

	return c, nil
}

// Create new todo task
func (s *todoServiceServer) Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {
	// check if the API version requested by client is suppoerted by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// get SQL Connection from pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}

	defer c.Close()

	reminder, err := ptypes.Timestamp(req.Todo.Reminder)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Reminder field has invalid format -> "+err.Error())
	}

	// insert Todo entity data
	query := `INSERT INTO todo(title, description, reminder) VALUES (?, ?, ?)`
	res, err := c.ExecContext(ctx, query, req.Todo.Title, req.Todo.Description, reminder)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to insert into todo -> "+err.Error())
	}

	// get ID of creates Todo
	id, err := res.LastInsertId()
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve id for created Todo -> "+err.Error())
	}

	return &CreateResponse{
		Api: apiVersion,
		Id:  id,
	}, nil
}

// Read todo task
func (s *todoServiceServer) Read(ctx context.Context, req *ReadRequest) (*ReadResponse, error) {
	// check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// get SQL connection from pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}

	defer c.Close()

	// query Todo by ID
	query := `SELECT id, title, description, reminder FROM todo where id = ?`
	rows, err := c.QueryContext(ctx, query, req.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to select from Todo -> "+err.Error())
	}

	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, status.Error(codes.Unknown, "failed to retrieve data from Todo -> "+err.Error())
		}
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Todo with ID='%d' is not found"))
	}

	// get Todo Data
	var td Todo
	var reminder time.Time

	if err := rows.Scan(
		&td.Id,
		&td.Title,
		&td.Description,
		&reminder,
	); err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("Found multiple Todo rows with ID='%d'", req.Id))
	}

	td.Reminder, err = ptypes.TimestampProto(reminder)
	if err != nil {
		return nil, status.Error(codes.Unknown, "reminder field has invalid format -> "+err.Error())
	}

	if rows.Next() {
		return nil, err
	}

	return &ReadResponse{
		Api:  apiVersion,
		Todo: &td,
	}, nil
}

// Update todo task
func (s *todoServiceServer) Update(ctx context.Context, req *UpdateRequest) (*UpdateResponse, error) {
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// get SQL connection from pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	reminder, err := ptypes.Timestamp(req.Todo.Reminder)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Reminder field has invalid format -> "+err.Error())
	}

	// update todo
	query := `UPDATE todo SET title = ?, description = ?, reminder = ? WHERE id = ?`
	res, err := c.ExecContext(
		ctx,
		query,
		req.Todo.Title,
		req.Todo.Description,
		reminder,
		req.Todo.Id,
	)

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to retrieve rows affected value ->"+err.Error())
	}

	return &UpdateResponse{
		Api:     apiVersion,
		Updated: rows,
	}, nil
}

// Delete todo task
func (s *todoServiceServer) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// get SQL Connection from pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// delete todo
	query := "DELETE FROM todo WHERE id = ?"
	res, err := c.ExecContext(ctx, query, req.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to delete Todo ->"+err.Error())
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to retrieve rows affected value -> "+err.Error())
	}

	if rows == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Todo with ID = '%d' is not found", req.Id))
	}

	return &DeleteResponse{
		Api:     apiVersion,
		Deleted: rows,
	}, nil
}

// Read all todo tasks
func (s *todoServiceServer) ReadAll(ctx context.Context, req *ReadAllRequest) (*ReadAllResponse, error) {
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// get SQL Connection from pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// get Todo Lost
	query := `SELECT id, title, description, reminder FROM todo`
	rows, err := c.QueryContext(ctx, query)

	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to select from todo -> "+err.Error())
	}
	defer rows.Close()

	var reminder time.Time
	list := []*Todo{}

	for rows.Next() {
		td := new(Todo)
		if err := rows.Scan(
			&td.Id,
			&td.Title,
			&td.Description,
			&reminder,
		); err != nil {
			return nil, status.Error(codes.Unknown, "Failed to retrieve field values from Todo -> "+err.Error())
		}

		td.Reminder, err = ptypes.TimestampProto(reminder)
		if err != nil {
			return nil, status.Error(codes.Unknown, "reminder field has invalid format -> "+err.Error())
		}
		list = append(list, td)
	}

	if err := rows.Err(); err != nil {
		return nil, status.Error(codes.Unknown, "Failed to retrieve data from Todo"+err.Error())
	}

	return &ReadAllResponse{
		Api:   apiVersion,
		Todos: list,
	}, nil
}
