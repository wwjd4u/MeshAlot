package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)
const Version="0.0.1-dev"
type Client struct{baseURL,nodeID,token string;httpClient *http.Client}
func NewClient(baseURL,nodeID,token string)*Client{return &Client{strings.TrimRight(baseURL,"/"),nodeID,token,&http.Client{Timeout:10*time.Second}}}
func(c *Client)Enroll(ctx context.Context)error{return c.post(ctx,"/v1/enroll",protocol.EnrollRequest{NodeID:c.nodeID,EnrollmentToken:c.token,AgentVersion:Version},201)}
func(c *Client)Heartbeat(ctx context.Context)error{return c.post(ctx,"/v1/heartbeat",protocol.HeartbeatRequest{NodeID:c.nodeID,ObservedAt:time.Now().UTC(),Mode:"available"},204)}
func(c *Client)post(ctx context.Context,path string,value any,expected int)error{
	body,err:=json.Marshal(value);if err!=nil{return err}
	req,err:=http.NewRequestWithContext(ctx,"POST",c.baseURL+path,bytes.NewReader(body));if err!=nil{return err}
	req.Header.Set("Content-Type","application/json")
	req.Header.Set("Authorization","Bearer "+c.token)
	resp,err:=c.httpClient.Do(req);if err!=nil{return err};defer resp.Body.Close()
	if resp.StatusCode!=expected{return fmt.Errorf("%s returned %s",path,resp.Status)};return nil
}
