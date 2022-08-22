package restclient

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

//RestClient - interface, declaring rest client methods
type RestClient interface {
	Post(url string, rqBody, successBody interface{}) error
	Get(url string, successBody, errorBody interface{}) error
	Delete(url string, rqBody, successBody interface{}) error
	UploadFile(url string, filePath string, successBody interface{}) error
	DownloadFile(url string, targetPath string, reqBody interface{}) error
}

//RestClientImpl - implementation of RestClient
type RestClientImpl struct {
	client   *resty.Client
	jwtToken string
}

type errorResponse struct {
	Status   int    `json:"errorcode"`
	ErrorMsg string `json:"message"`
}

//NewRestClientImpl - used for creation of instances of RestClientImpl
func NewRestClientImpl(jwtToken string) *RestClientImpl {
	return &RestClientImpl{
		client:   resty.New(),
		jwtToken: jwtToken,
	}
}

//Post - creation of resources
func (i *RestClientImpl) Post(url string, rqBody, successBody interface{}) error {
	errorBody := errorResponse{}
	resp, err := i.basicRequest(successBody, &errorBody).
		SetBody(rqBody).
		Post(url)

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusCreated {
		return fmt.Errorf("Problem with Post request. Reason: %s", errorBody.ErrorMsg)
	}
	return nil
}

//Get - retrieval of resources
func (i *RestClientImpl) Get(url string, successBody interface{}) error {
	errorBody := errorResponse{}
	resp, err := i.basicRequest(successBody, &errorBody).
		Get(url)

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("Problem with Get request. Reason: %s", errorBody.ErrorMsg)
	}
	return nil
}

//Delete - deletion of resources
func (i *RestClientImpl) Delete(url string, reqBody, successBody interface{}) error {
	errorBody := errorResponse{}
	resp, err := i.basicRequest(successBody, &errorBody).
		SetBody(reqBody).
		Delete(url)

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("Problem with Delete request. Reason: %s", errorBody.ErrorMsg)
	}
	return nil
}

//UploadFile - similar to POST, but it uses form-data to include the payload (file)
func (i *RestClientImpl) UploadFile(url string, filePath string, successBody interface{}) error {
	errorBody := errorResponse{}
	req := i.client.R().
		SetFile("file", filePath).
		SetError(&errorBody)

	if i.jwtToken != "" {
		req.SetAuthToken(i.jwtToken)
	}

	if successBody != nil {
		req.SetResult(successBody)
	}

	resp, err := req.Post(url)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusCreated {
		return fmt.Errorf("Problem with the Upload file request. Reason: %s", errorBody.ErrorMsg)
	}

	return nil
}

//DownloadFile - similar to GET, but it requires the target location where the file will be downloaded
func (i *RestClientImpl) DownloadFile(url string, targetPath string) error {
	errorBody := errorResponse{}
	req := i.client.R().
		SetOutput(targetPath).
		SetError(&errorBody)

	if i.jwtToken != "" {
		req.SetAuthToken(i.jwtToken)
	}

	resp, err := req.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("Problem with the Download file request. Reason: %s", errorBody.ErrorMsg)
	}

	return nil
}

func (i *RestClientImpl) basicRequest(successBody, errorBody interface{}) *resty.Request {
	req := i.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetError(&errorBody)

	if i.jwtToken != "" {
		req.SetAuthToken(i.jwtToken)
	}

	if successBody != nil {
		req.SetResult(&successBody)
	}

	return req
}
