// matrix-skype - A Matrix-WhatsApp puppeting bridge.
// Copyright (C) 2019 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package skypeExt

import (
	"encoding/json"
	skype "github.com/kelaresg/go-skypeapi"
	"strings"
)

type ChatUpdateCommand string

const (
	ChatUpdateCommandAction ChatUpdateCommand = "action"
)

type ChatUpdate struct {
	JID     string            `json:"id"`
	Command ChatUpdateCommand `json:"cmd"`
	Data    ChatUpdateData    `json:"data"`
}

type ChatActionType string

const (
	ChatActionNameChange  ChatActionType = "subject"
	ChatActionAddTopic    ChatActionType = "desc_add"
	ChatActionRemoveTopic ChatActionType = "desc_remove"
	ChatActionRestrict    ChatActionType = "restrict"
	ChatActionAnnounce    ChatActionType = "announce"
	ChatActionPromote     ChatActionType = "promote"
	ChatActionDemote      ChatActionType = "demote"
	ChatActionRemove      ChatActionType = "remove"
	ChatActionAdd         ChatActionType = "add"
	ChatActionIntroduce   ChatActionType = "introduce"
	ChatActionCreate      ChatActionType = "create"
	ChatActionThread      ChatActionType = "Thread"
)

const (
	ChatTopicUpdate   ChatActionType = "ThreadActivity/TopicUpdate"
	ChatPictureUpdate ChatActionType = "ThreadActivity/PictureUpdate"
	ChatMemberAdd     ChatActionType = "ThreadActivity/AddMember"
	ChatMemberDelete  ChatActionType = "ThreadActivity/DeleteMember"
)

type ChatUpdateData struct {
	Action    ChatActionType
	SenderJID string

	NameChange struct {
		Name  string `json:"subject"`
		SetAt int64  `json:"s_t"`
		SetBy string `json:"s_o"`
	}

	AddTopic struct {
		Topic string `json:"desc"`
		ID    string `json:"descId"`
		SetAt int64  `json:"descTime"`
	}

	RemoveTopic struct {
		ID string `json:"descId"`
	}

	Restrict bool

	Announce bool

	PermissionChange struct {
		JIDs []string `json:"participants"`
	}

	MemberAction struct {
		JIDs []string `json:"participants"`
	}

	Create struct {
		Creation    int64    `json:"creation"`
		Name        string   `json:"subject"`
		SetAt       int64    `json:"s_t"`
		SetBy       string   `json:"s_o"`
		Admins      []string `json:"admins"`
		SuperAdmins []string `json:"superadmins"`
		Regulars    []string `json:"regulars"`
	}
}

func (cud *ChatUpdateData) UnmarshalJSON(data []byte) error {
	var arr []json.RawMessage
	err := json.Unmarshal(data, &arr)
	if err != nil {
		return err
	} else if len(arr) < 3 {
		return nil
	}

	err = json.Unmarshal(arr[0], &cud.Action)
	if err != nil {
		return err
	}

	err = json.Unmarshal(arr[1], &cud.SenderJID)
	if err != nil {
		return err
	}
	cud.SenderJID = strings.Replace(cud.SenderJID, OldUserSuffix, NewUserSuffix, 1)

	var unmarshalTo interface{}
	switch cud.Action {
	case ChatActionNameChange:
		unmarshalTo = &cud.NameChange
	case ChatActionAddTopic:
		unmarshalTo = &cud.AddTopic
	case ChatActionRemoveTopic:
		unmarshalTo = &cud.RemoveTopic
	case ChatActionRestrict:
		unmarshalTo = &cud.Restrict
	case ChatActionAnnounce:
		unmarshalTo = &cud.Announce
	case ChatActionPromote, ChatActionDemote:
		unmarshalTo = &cud.PermissionChange
	case ChatActionAdd, ChatActionRemove:
		unmarshalTo = &cud.MemberAction
	case ChatActionCreate:
		unmarshalTo = &cud.Create
	default:
		return nil
	}
	err = json.Unmarshal(arr[2], unmarshalTo)
	if err != nil {
		return err
	}
	cud.NameChange.SetBy = strings.Replace(cud.NameChange.SetBy, OldUserSuffix, NewUserSuffix, 1)
	for index, jid := range cud.PermissionChange.JIDs {
		cud.PermissionChange.JIDs[index] = strings.Replace(jid, OldUserSuffix, NewUserSuffix, 1)
	}
	return nil
}

type ChatUpdateHandler interface {
	skype.Handler
	HandleChatUpdate(ChatUpdate)
}

func (ext *ExtendedConn) handleMessageChatUpdate(message []byte) {
	var event ChatUpdate
	err := json.Unmarshal(message, &event)
	if err != nil {
		ext.jsonParseError(err)
		return
	}
	event.JID = strings.Replace(event.JID, OldUserSuffix, NewUserSuffix, 1)
	for _, handler := range ext.handlers {
		chatUpdateHandler, ok := handler.(ChatUpdateHandler)
		if !ok {
			continue
		}

		if ext.shouldCallSynchronously(chatUpdateHandler) {
			chatUpdateHandler.HandleChatUpdate(event)
		} else {
			go chatUpdateHandler.HandleChatUpdate(event)
		}
	}
}
