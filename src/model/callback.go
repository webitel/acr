/**
 * Created by I. Navrotskyj on 30.10.17.
 */

package model

type CallbackMember struct {
	Id int
}

type ExistsCallbackMemberRequest struct {
	Number *string `json:"number"`
	Done   *bool   `json:"done"`
}
