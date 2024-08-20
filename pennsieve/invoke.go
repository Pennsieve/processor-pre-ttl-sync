package pennsieve

import (
	"fmt"
	"github.com/pennsieve/processor-pre-ttl-sync/util"
	"io"
	"net/http"
)

type Session struct {
	Token    string
	APIHost  string
	API2Host string
}

func NewSession(sessionToken, apiHost, api2Host string) *Session {
	return &Session{
		Token:    sessionToken,
		APIHost:  apiHost,
		API2Host: api2Host}
}

func (s *Session) newPennsieveRequest(method string, url string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating GET %s request: %w", url, err)
	}
	request.Header.Add("accept", "application/json")
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.Token))
	return request, nil
}

func (s *Session) InvokePennsieve(method string, url string, body io.Reader) (*http.Response, error) {

	req, err := s.newPennsieveRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating %s %s request: %w", method, url, err)
	}
	return util.Invoke(req)
}
