package container

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

// ContainerWaitOKBody ContainerWaitResponse
//
// OK response to ContainerWait operation
// swagger:model ContainerWaitOKBody
type ContainerWaitOKBody struct {

	// error
	// Required: true
	Error *ContainerWaitOKBodyError `json:"Error"`

	// Exit code of the container
	// Required: true
	StatusCode int64 `json:"StatusCode"`
}
