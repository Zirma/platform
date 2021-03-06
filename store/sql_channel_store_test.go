// Copyright (c) 2015 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package store

import (
	"testing"
	"time"

	"github.com/mattermost/platform/model"
)

func TestChannelStoreSave(t *testing.T) {
	Setup()

	teamId := model.NewId()

	o1 := model.Channel{}
	o1.TeamId = teamId
	o1.DisplayName = "Name"
	o1.Name = "a" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN

	if err := (<-store.Channel().Save(&o1)).Err; err != nil {
		t.Fatal("couldn't save item", err)
	}

	if err := (<-store.Channel().Save(&o1)).Err; err == nil {
		t.Fatal("shouldn't be able to update from save")
	}

	o1.Id = ""
	if err := (<-store.Channel().Save(&o1)).Err; err == nil {
		t.Fatal("should be unique name")
	}

	o1.Id = ""
	o1.Name = "a" + model.NewId() + "b"
	o1.Type = model.CHANNEL_DIRECT
	if err := (<-store.Channel().Save(&o1)).Err; err == nil {
		t.Fatal("Should not be able to save direct channel")
	}
}

func TestChannelStoreSaveDirectChannel(t *testing.T) {
	Setup()

	teamId := model.NewId()

	o1 := model.Channel{}
	o1.TeamId = teamId
	o1.DisplayName = "Name"
	o1.Name = "a" + model.NewId() + "b"
	o1.Type = model.CHANNEL_DIRECT

	u1 := &model.User{}
	u1.Email = model.NewId()
	u1.Nickname = model.NewId()
	Must(store.User().Save(u1))
	Must(store.Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u1.Id}))

	u2 := &model.User{}
	u2.Email = model.NewId()
	u2.Nickname = model.NewId()
	Must(store.User().Save(u2))
	Must(store.Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u2.Id}))

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = u1.Id
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()

	m2 := model.ChannelMember{}
	m2.ChannelId = o1.Id
	m2.UserId = u2.Id
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()

	if err := (<-store.Channel().SaveDirectChannel(&o1, &m1, &m2)).Err; err != nil {
		t.Fatal("couldn't save direct channel", err)
	}

	members := (<-store.Channel().GetMembers(o1.Id, 0, 100)).Data.(*model.ChannelMembers)
	if len(*members) != 2 {
		t.Fatal("should have saved 2 members")
	}

	if err := (<-store.Channel().SaveDirectChannel(&o1, &m1, &m2)).Err; err == nil {
		t.Fatal("shouldn't be able to update from save")
	}

	// Attempt to save a direct channel that already exists
	o1a := model.Channel{
		TeamId:      o1.TeamId,
		DisplayName: o1.DisplayName,
		Name:        o1.Name,
		Type:        o1.Type,
	}

	if result := <-store.Channel().SaveDirectChannel(&o1a, &m1, &m2); result.Err == nil {
		t.Fatal("should've failed to save a duplicate direct channel")
	} else if result.Err.Id != CHANNEL_EXISTS_ERROR {
		t.Fatal("should've returned CHANNEL_EXISTS_ERROR")
	} else if returned := result.Data.(*model.Channel); returned.Id != o1.Id {
		t.Fatal("should've returned original channel when saving a duplicate direct channel")
	}

	// Attempt to save a non-direct channel
	o1.Id = ""
	o1.Name = "a" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	if err := (<-store.Channel().SaveDirectChannel(&o1, &m1, &m2)).Err; err == nil {
		t.Fatal("Should not be able to save non-direct channel")
	}
}

func TestChannelStoreCreateDirectChannel(t *testing.T) {
	Setup()

	u1 := &model.User{}
	u1.Email = model.NewId()
	u1.Nickname = model.NewId()
	Must(store.User().Save(u1))
	Must(store.Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u1.Id}))

	u2 := &model.User{}
	u2.Email = model.NewId()
	u2.Nickname = model.NewId()
	Must(store.User().Save(u2))
	Must(store.Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u2.Id}))

	res := <-store.Channel().CreateDirectChannel(u1.Id, u2.Id)
	if res.Err != nil {
		t.Fatal("couldn't create direct channel", res.Err)
	}

	c1 := res.Data.(*model.Channel)

	members := (<-store.Channel().GetMembers(c1.Id, 0, 100)).Data.(*model.ChannelMembers)
	if len(*members) != 2 {
		t.Fatal("should have saved 2 members")
	}
}

func TestChannelStoreUpdate(t *testing.T) {
	Setup()

	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "Name"
	o1.Name = "a" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN

	if err := (<-store.Channel().Save(&o1)).Err; err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	if err := (<-store.Channel().Update(&o1)).Err; err != nil {
		t.Fatal(err)
	}

	o1.Id = "missing"
	if err := (<-store.Channel().Update(&o1)).Err; err == nil {
		t.Fatal("Update should have failed because of missing key")
	}

	o1.Id = model.NewId()
	if err := (<-store.Channel().Update(&o1)).Err; err == nil {
		t.Fatal("Update should have faile because id change")
	}
}

func TestChannelStoreGet(t *testing.T) {
	Setup()

	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "Name"
	o1.Name = "a" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o1))

	if r1 := <-store.Channel().Get(o1.Id, false); r1.Err != nil {
		t.Fatal(r1.Err)
	} else {
		if r1.Data.(*model.Channel).ToJson() != o1.ToJson() {
			t.Fatal("invalid returned channel")
		}
	}

	if err := (<-store.Channel().Get("", false)).Err; err == nil {
		t.Fatal("Missing id should have failed")
	}

	u1 := &model.User{}
	u1.Email = model.NewId()
	u1.Nickname = model.NewId()
	Must(store.User().Save(u1))
	Must(store.Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u1.Id}))

	u2 := model.User{}
	u2.Email = model.NewId()
	u2.Nickname = model.NewId()
	Must(store.User().Save(&u2))
	Must(store.Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u2.Id}))

	o2 := model.Channel{}
	o2.TeamId = model.NewId()
	o2.DisplayName = "Direct Name"
	o2.Name = "a" + model.NewId() + "b"
	o2.Type = model.CHANNEL_DIRECT

	m1 := model.ChannelMember{}
	m1.ChannelId = o2.Id
	m1.UserId = u1.Id
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()

	m2 := model.ChannelMember{}
	m2.ChannelId = o2.Id
	m2.UserId = u2.Id
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()

	Must(store.Channel().SaveDirectChannel(&o2, &m1, &m2))

	if r2 := <-store.Channel().Get(o2.Id, false); r2.Err != nil {
		t.Fatal(r2.Err)
	} else {
		if r2.Data.(*model.Channel).ToJson() != o2.ToJson() {
			t.Fatal("invalid returned channel")
		}
	}

	if r4 := <-store.Channel().Get(o2.Id, true); r4.Err != nil {
		t.Fatal(r4.Err)
	} else {
		if r4.Data.(*model.Channel).ToJson() != o2.ToJson() {
			t.Fatal("invalid returned channel")
		}
	}

	if r3 := <-store.Channel().GetAll(o1.TeamId); r3.Err != nil {
		t.Fatal(r3.Err)
	} else {
		channels := r3.Data.([]*model.Channel)
		if len(channels) == 0 {
			t.Fatal("too little")
		}
	}
}

func TestChannelStoreGetForPost(t *testing.T) {
	Setup()

	o1 := Must(store.Channel().Save(&model.Channel{
		TeamId:      model.NewId(),
		DisplayName: "Name",
		Name:        "a" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	})).(*model.Channel)

	p1 := Must(store.Post().Save(&model.Post{
		UserId:    model.NewId(),
		ChannelId: o1.Id,
		Message:   "test",
	})).(*model.Post)

	if r1 := <-store.Channel().GetForPost(p1.Id); r1.Err != nil {
		t.Fatal(r1.Err)
	} else if r1.Data.(*model.Channel).Id != o1.Id {
		t.Fatal("incorrect channel returned")
	}
}

func TestChannelStoreDelete(t *testing.T) {
	Setup()

	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "Channel1"
	o1.Name = "a" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o1))

	o2 := model.Channel{}
	o2.TeamId = o1.TeamId
	o2.DisplayName = "Channel2"
	o2.Name = "a" + model.NewId() + "b"
	o2.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o2))

	o3 := model.Channel{}
	o3.TeamId = o1.TeamId
	o3.DisplayName = "Channel3"
	o3.Name = "a" + model.NewId() + "b"
	o3.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o3))

	o4 := model.Channel{}
	o4.TeamId = o1.TeamId
	o4.DisplayName = "Channel4"
	o4.Name = "a" + model.NewId() + "b"
	o4.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o4))

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = model.NewId()
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&m1))

	m2 := model.ChannelMember{}
	m2.ChannelId = o2.Id
	m2.UserId = m1.UserId
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&m2))

	if r := <-store.Channel().Delete(o1.Id, model.GetMillis()); r.Err != nil {
		t.Fatal(r.Err)
	}

	if r := <-store.Channel().Get(o1.Id, false); r.Data.(*model.Channel).DeleteAt == 0 {
		t.Fatal("should have been deleted")
	}

	if r := <-store.Channel().Delete(o3.Id, model.GetMillis()); r.Err != nil {
		t.Fatal(r.Err)
	}

	cresult := <-store.Channel().GetChannels(o1.TeamId, m1.UserId)
	list := cresult.Data.(*model.ChannelList)

	if len(*list) != 1 {
		t.Fatal("invalid number of channels")
	}

	cresult = <-store.Channel().GetMoreChannels(o1.TeamId, m1.UserId, 0, 100)
	list = cresult.Data.(*model.ChannelList)

	if len(*list) != 1 {
		t.Fatal("invalid number of channels")
	}

	<-store.Channel().PermanentDelete(o2.Id)

	cresult = <-store.Channel().GetChannels(o1.TeamId, m1.UserId)
	t.Log(cresult.Err)
	if cresult.Err.Id != "store.sql_channel.get_channels.not_found.app_error" {
		t.Fatal("no channels should be found")
	}
}

func TestChannelStoreGetByName(t *testing.T) {
	Setup()

	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "Name"
	o1.Name = "a" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o1))

	r1 := <-store.Channel().GetByName(o1.TeamId, o1.Name, true)
	if r1.Err != nil {
		t.Fatal(r1.Err)
	} else {
		if r1.Data.(*model.Channel).ToJson() != o1.ToJson() {
			t.Fatal("invalid returned channel")
		}
	}

	if err := (<-store.Channel().GetByName(o1.TeamId, "", true)).Err; err == nil {
		t.Fatal("Missing id should have failed")
	}

	if r1 := <-store.Channel().GetByName(o1.TeamId, o1.Name, false); r1.Err != nil {
		t.Fatal(r1.Err)
	} else {
		if r1.Data.(*model.Channel).ToJson() != o1.ToJson() {
			t.Fatal("invalid returned channel")
		}
	}

	if err := (<-store.Channel().GetByName(o1.TeamId, "", false)).Err; err == nil {
		t.Fatal("Missing id should have failed")
	}

	Must(store.Channel().Delete(r1.Data.(*model.Channel).Id, model.GetMillis()))

	if err := (<-store.Channel().GetByName(o1.TeamId, "", false)).Err; err == nil {
		t.Fatal("Deleted channel should not be returned by GetByName()")
	}
}

func TestChannelStoreGetDeletedByName(t *testing.T) {
	Setup()

	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "Name"
	o1.Name = "a" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	o1.DeleteAt = model.GetMillis()
	Must(store.Channel().Save(&o1))

	if r1 := <-store.Channel().GetDeletedByName(o1.TeamId, o1.Name); r1.Err != nil {
		t.Fatal(r1.Err)
	} else {
		if r1.Data.(*model.Channel).ToJson() != o1.ToJson() {
			t.Fatal("invalid returned channel")
		}
	}

	if err := (<-store.Channel().GetDeletedByName(o1.TeamId, "")).Err; err == nil {
		t.Fatal("Missing id should have failed")
	}
}

func TestChannelMemberStore(t *testing.T) {
	Setup()

	c1 := model.Channel{}
	c1.TeamId = model.NewId()
	c1.DisplayName = "NameName"
	c1.Name = "a" + model.NewId() + "b"
	c1.Type = model.CHANNEL_OPEN
	c1 = *Must(store.Channel().Save(&c1)).(*model.Channel)

	c1t1 := (<-store.Channel().Get(c1.Id, false)).Data.(*model.Channel)
	t1 := c1t1.ExtraUpdateAt

	u1 := model.User{}
	u1.Email = model.NewId()
	u1.Nickname = model.NewId()
	Must(store.User().Save(&u1))
	Must(store.Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u1.Id}))

	u2 := model.User{}
	u2.Email = model.NewId()
	u2.Nickname = model.NewId()
	Must(store.User().Save(&u2))
	Must(store.Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u2.Id}))

	o1 := model.ChannelMember{}
	o1.ChannelId = c1.Id
	o1.UserId = u1.Id
	o1.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&o1))

	o2 := model.ChannelMember{}
	o2.ChannelId = c1.Id
	o2.UserId = u2.Id
	o2.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&o2))

	c1t2 := (<-store.Channel().Get(c1.Id, false)).Data.(*model.Channel)
	t2 := c1t2.ExtraUpdateAt

	if t2 <= t1 {
		t.Fatal("Member update time incorrect")
	}

	count := (<-store.Channel().GetMemberCount(o1.ChannelId, true)).Data.(int64)
	if count != 2 {
		t.Fatal("should have saved 2 members")
	}

	count = (<-store.Channel().GetMemberCount(o1.ChannelId, true)).Data.(int64)
	if count != 2 {
		t.Fatal("should have saved 2 members")
	}

	if store.Channel().GetMemberCountFromCache(o1.ChannelId) != 2 {
		t.Fatal("should have saved 2 members")
	}

	if store.Channel().GetMemberCountFromCache("junk") != 0 {
		t.Fatal("should have saved 0 members")
	}

	count = (<-store.Channel().GetMemberCount(o1.ChannelId, false)).Data.(int64)
	if count != 2 {
		t.Fatal("should have saved 2 members")
	}

	Must(store.Channel().RemoveMember(o2.ChannelId, o2.UserId))

	count = (<-store.Channel().GetMemberCount(o1.ChannelId, false)).Data.(int64)
	if count != 1 {
		t.Fatal("should have removed 1 member")
	}

	c1t3 := (<-store.Channel().Get(c1.Id, false)).Data.(*model.Channel)
	t3 := c1t3.ExtraUpdateAt

	if t3 <= t2 || t3 <= t1 {
		t.Fatal("Member update time incorrect on delete")
	}

	member := (<-store.Channel().GetMember(o1.ChannelId, o1.UserId, false)).Data.(*model.ChannelMember)
	if member.ChannelId != o1.ChannelId {
		t.Fatal("should have go member")
	}

	if err := (<-store.Channel().SaveMember(&o1)).Err; err == nil {
		t.Fatal("Should have been a duplicate")
	}

	c1t4 := (<-store.Channel().Get(c1.Id, false)).Data.(*model.Channel)
	t4 := c1t4.ExtraUpdateAt
	if t4 != t3 {
		t.Fatal("Should not update time upon failure")
	}
}

func TestChannelDeleteMemberStore(t *testing.T) {
	Setup()

	c1 := model.Channel{}
	c1.TeamId = model.NewId()
	c1.DisplayName = "NameName"
	c1.Name = "a" + model.NewId() + "b"
	c1.Type = model.CHANNEL_OPEN
	c1 = *Must(store.Channel().Save(&c1)).(*model.Channel)

	c1t1 := (<-store.Channel().Get(c1.Id, false)).Data.(*model.Channel)
	t1 := c1t1.ExtraUpdateAt

	u1 := model.User{}
	u1.Email = model.NewId()
	u1.Nickname = model.NewId()
	Must(store.User().Save(&u1))
	Must(store.Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u1.Id}))

	u2 := model.User{}
	u2.Email = model.NewId()
	u2.Nickname = model.NewId()
	Must(store.User().Save(&u2))
	Must(store.Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u2.Id}))

	o1 := model.ChannelMember{}
	o1.ChannelId = c1.Id
	o1.UserId = u1.Id
	o1.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&o1))

	o2 := model.ChannelMember{}
	o2.ChannelId = c1.Id
	o2.UserId = u2.Id
	o2.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&o2))

	c1t2 := (<-store.Channel().Get(c1.Id, false)).Data.(*model.Channel)
	t2 := c1t2.ExtraUpdateAt

	if t2 <= t1 {
		t.Fatal("Member update time incorrect")
	}

	count := (<-store.Channel().GetMemberCount(o1.ChannelId, false)).Data.(int64)
	if count != 2 {
		t.Fatal("should have saved 2 members")
	}

	Must(store.Channel().PermanentDeleteMembersByUser(o2.UserId))

	count = (<-store.Channel().GetMemberCount(o1.ChannelId, false)).Data.(int64)
	if count != 1 {
		t.Fatal("should have removed 1 member")
	}

	if r1 := <-store.Channel().PermanentDeleteMembersByChannel(o1.ChannelId); r1.Err != nil {
		t.Fatal(r1.Err)
	}

	count = (<-store.Channel().GetMemberCount(o1.ChannelId, false)).Data.(int64)
	if count != 0 {
		t.Fatal("should have removed all members")
	}
}

func TestChannelStoreGetChannels(t *testing.T) {
	Setup()

	o2 := model.Channel{}
	o2.TeamId = model.NewId()
	o2.DisplayName = "Channel2"
	o2.Name = "a" + model.NewId() + "b"
	o2.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o2))

	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "Channel1"
	o1.Name = "a" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o1))

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = model.NewId()
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&m1))

	m2 := model.ChannelMember{}
	m2.ChannelId = o1.Id
	m2.UserId = model.NewId()
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&m2))

	m3 := model.ChannelMember{}
	m3.ChannelId = o2.Id
	m3.UserId = model.NewId()
	m3.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&m3))

	cresult := <-store.Channel().GetChannels(o1.TeamId, m1.UserId)
	list := cresult.Data.(*model.ChannelList)

	if (*list)[0].Id != o1.Id {
		t.Fatal("missing channel")
	}

	acresult := <-store.Channel().GetAllChannelMembersForUser(m1.UserId, false)
	ids := acresult.Data.(map[string]string)
	if _, ok := ids[o1.Id]; !ok {
		t.Fatal("missing channel")
	}

	acresult2 := <-store.Channel().GetAllChannelMembersForUser(m1.UserId, true)
	ids2 := acresult2.Data.(map[string]string)
	if _, ok := ids2[o1.Id]; !ok {
		t.Fatal("missing channel")
	}

	acresult3 := <-store.Channel().GetAllChannelMembersForUser(m1.UserId, true)
	ids3 := acresult3.Data.(map[string]string)
	if _, ok := ids3[o1.Id]; !ok {
		t.Fatal("missing channel")
	}

	if !store.Channel().IsUserInChannelUseCache(m1.UserId, o1.Id) {
		t.Fatal("missing channel")
	}

	if store.Channel().IsUserInChannelUseCache(m1.UserId, o2.Id) {
		t.Fatal("missing channel")
	}

	if store.Channel().IsUserInChannelUseCache(m1.UserId, "blahblah") {
		t.Fatal("missing channel")
	}

	if store.Channel().IsUserInChannelUseCache("blahblah", "blahblah") {
		t.Fatal("missing channel")
	}

	store.Channel().InvalidateAllChannelMembersForUser(m1.UserId)
}

func TestChannelStoreGetMoreChannels(t *testing.T) {
	Setup()

	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "Channel1"
	o1.Name = "a" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o1))

	o2 := model.Channel{}
	o2.TeamId = model.NewId()
	o2.DisplayName = "Channel2"
	o2.Name = "a" + model.NewId() + "b"
	o2.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o2))

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = model.NewId()
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&m1))

	m2 := model.ChannelMember{}
	m2.ChannelId = o1.Id
	m2.UserId = model.NewId()
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&m2))

	m3 := model.ChannelMember{}
	m3.ChannelId = o2.Id
	m3.UserId = model.NewId()
	m3.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&m3))

	o3 := model.Channel{}
	o3.TeamId = o1.TeamId
	o3.DisplayName = "ChannelA"
	o3.Name = "a" + model.NewId() + "b"
	o3.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o3))

	o4 := model.Channel{}
	o4.TeamId = o1.TeamId
	o4.DisplayName = "ChannelB"
	o4.Name = "a" + model.NewId() + "b"
	o4.Type = model.CHANNEL_PRIVATE
	Must(store.Channel().Save(&o4))

	o5 := model.Channel{}
	o5.TeamId = o1.TeamId
	o5.DisplayName = "ChannelC"
	o5.Name = "a" + model.NewId() + "b"
	o5.Type = model.CHANNEL_PRIVATE
	Must(store.Channel().Save(&o5))

	cresult := <-store.Channel().GetMoreChannels(o1.TeamId, m1.UserId, 0, 100)
	if cresult.Err != nil {
		t.Fatal(cresult.Err)
	}
	list := cresult.Data.(*model.ChannelList)

	if len(*list) != 1 {
		t.Fatal("wrong list")
	}

	if (*list)[0].Name != o3.Name {
		t.Fatal("missing channel")
	}

	o6 := model.Channel{}
	o6.TeamId = o1.TeamId
	o6.DisplayName = "ChannelA"
	o6.Name = "a" + model.NewId() + "b"
	o6.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o6))

	cresult = <-store.Channel().GetMoreChannels(o1.TeamId, m1.UserId, 0, 100)
	list = cresult.Data.(*model.ChannelList)

	if len(*list) != 2 {
		t.Fatal("wrong list length")
	}

	cresult = <-store.Channel().GetMoreChannels(o1.TeamId, m1.UserId, 0, 1)
	list = cresult.Data.(*model.ChannelList)

	if len(*list) != 1 {
		t.Fatal("wrong list length")
	}

	cresult = <-store.Channel().GetMoreChannels(o1.TeamId, m1.UserId, 1, 1)
	list = cresult.Data.(*model.ChannelList)

	if len(*list) != 1 {
		t.Fatal("wrong list length")
	}

	if r1 := <-store.Channel().AnalyticsTypeCount(o1.TeamId, model.CHANNEL_OPEN); r1.Err != nil {
		t.Fatal(r1.Err)
	} else {
		if r1.Data.(int64) != 3 {
			t.Log(r1.Data)
			t.Fatal("wrong value")
		}
	}

	if r1 := <-store.Channel().AnalyticsTypeCount(o1.TeamId, model.CHANNEL_PRIVATE); r1.Err != nil {
		t.Fatal(r1.Err)
	} else {
		if r1.Data.(int64) != 2 {
			t.Log(r1.Data)
			t.Fatal("wrong value")
		}
	}
}

func TestChannelStoreGetChannelCounts(t *testing.T) {
	Setup()

	o2 := model.Channel{}
	o2.TeamId = model.NewId()
	o2.DisplayName = "Channel2"
	o2.Name = "a" + model.NewId() + "b"
	o2.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o2))

	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "Channel1"
	o1.Name = "a" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o1))

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = model.NewId()
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&m1))

	m2 := model.ChannelMember{}
	m2.ChannelId = o1.Id
	m2.UserId = model.NewId()
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&m2))

	m3 := model.ChannelMember{}
	m3.ChannelId = o2.Id
	m3.UserId = model.NewId()
	m3.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&m3))

	cresult := <-store.Channel().GetChannelCounts(o1.TeamId, m1.UserId)
	counts := cresult.Data.(*model.ChannelCounts)

	if len(counts.Counts) != 1 {
		t.Fatal("wrong number of counts")
	}

	if len(counts.UpdateTimes) != 1 {
		t.Fatal("wrong number of update times")
	}
}

func TestChannelStoreGetMembersForUser(t *testing.T) {
	Setup()

	t1 := model.Team{}
	t1.DisplayName = "Name"
	t1.Name = model.NewId()
	t1.Email = model.NewId() + "@nowhere.com"
	t1.Type = model.TEAM_OPEN
	Must(store.Team().Save(&t1))

	o1 := model.Channel{}
	o1.TeamId = t1.Id
	o1.DisplayName = "Channel1"
	o1.Name = "a" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o1))

	o2 := model.Channel{}
	o2.TeamId = o1.TeamId
	o2.DisplayName = "Channel2"
	o2.Name = "a" + model.NewId() + "b"
	o2.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o2))

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = model.NewId()
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&m1))

	m2 := model.ChannelMember{}
	m2.ChannelId = o2.Id
	m2.UserId = m1.UserId
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&m2))

	cresult := <-store.Channel().GetMembersForUser(o1.TeamId, m1.UserId)
	members := cresult.Data.(*model.ChannelMembers)

	// no unread messages
	if len(*members) != 2 {
		t.Fatal("wrong number of members")
	}
}

func TestChannelStoreUpdateLastViewedAt(t *testing.T) {
	Setup()

	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "Channel1"
	o1.Name = "a" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	o1.TotalMsgCount = 25
	Must(store.Channel().Save(&o1))

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = model.NewId()
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&m1))

	err := (<-store.Channel().UpdateLastViewedAt([]string{m1.ChannelId}, m1.UserId)).Err
	if err != nil {
		t.Fatal("failed to update", err)
	}

	err = (<-store.Channel().UpdateLastViewedAt([]string{m1.ChannelId}, "missing id")).Err
	if err != nil {
		t.Fatal("failed to update")
	}
}

func TestChannelStoreIncrementMentionCount(t *testing.T) {
	Setup()

	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "Channel1"
	o1.Name = "a" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	o1.TotalMsgCount = 25
	Must(store.Channel().Save(&o1))

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = model.NewId()
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&m1))

	err := (<-store.Channel().IncrementMentionCount(m1.ChannelId, m1.UserId)).Err
	if err != nil {
		t.Fatal("failed to update")
	}

	err = (<-store.Channel().IncrementMentionCount(m1.ChannelId, "missing id")).Err
	if err != nil {
		t.Fatal("failed to update")
	}

	err = (<-store.Channel().IncrementMentionCount("missing id", m1.UserId)).Err
	if err != nil {
		t.Fatal("failed to update")
	}

	err = (<-store.Channel().IncrementMentionCount("missing id", "missing id")).Err
	if err != nil {
		t.Fatal("failed to update")
	}
}

func TestGetMember(t *testing.T) {
	Setup()

	userId := model.NewId()

	c1 := &model.Channel{
		TeamId:      model.NewId(),
		DisplayName: model.NewId(),
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
	}
	Must(store.Channel().Save(c1))

	c2 := &model.Channel{
		TeamId:      c1.TeamId,
		DisplayName: model.NewId(),
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
	}
	Must(store.Channel().Save(c2))

	m1 := &model.ChannelMember{
		ChannelId:   c1.Id,
		UserId:      userId,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	Must(store.Channel().SaveMember(m1))

	m2 := &model.ChannelMember{
		ChannelId:   c2.Id,
		UserId:      userId,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	Must(store.Channel().SaveMember(m2))

	if result := <-store.Channel().GetMember(model.NewId(), userId, false); result.Err == nil {
		t.Fatal("should've failed to get member for non-existant channel")
	}

	if result := <-store.Channel().GetMember(c1.Id, model.NewId(), false); result.Err == nil {
		t.Fatal("should've failed to get member for non-existant user")
	}

	if result := <-store.Channel().GetMember(c1.Id, userId, false); result.Err != nil {
		t.Fatal("shouldn't have errored when getting member", result.Err)
	} else if member := result.Data.(*model.ChannelMember); member.ChannelId != c1.Id {
		t.Fatal("should've gotten member of channel 1")
	} else if member.UserId != userId {
		t.Fatal("should've gotten member for user")
	}

	if result := <-store.Channel().GetMember(c2.Id, userId, false); result.Err != nil {
		t.Fatal("shouldn't have errored when getting member", result.Err)
	} else if member := result.Data.(*model.ChannelMember); member.ChannelId != c2.Id {
		t.Fatal("should've gotten member of channel 2")
	} else if member.UserId != userId {
		t.Fatal("should've gotten member for user")
	}
}

func TestChannelStoreGetMemberForPost(t *testing.T) {
	Setup()

	o1 := Must(store.Channel().Save(&model.Channel{
		TeamId:      model.NewId(),
		DisplayName: "Name",
		Name:        "a" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	})).(*model.Channel)

	m1 := Must(store.Channel().SaveMember(&model.ChannelMember{
		ChannelId:   o1.Id,
		UserId:      model.NewId(),
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	})).(*model.ChannelMember)

	p1 := Must(store.Post().Save(&model.Post{
		UserId:    model.NewId(),
		ChannelId: o1.Id,
		Message:   "test",
	})).(*model.Post)

	if r1 := <-store.Channel().GetMemberForPost(p1.Id, m1.UserId); r1.Err != nil {
		t.Fatal(r1.Err)
	} else if r1.Data.(*model.ChannelMember).ToJson() != m1.ToJson() {
		t.Fatal("invalid returned channel member")
	}

	if r2 := <-store.Channel().GetMemberForPost(p1.Id, model.NewId()); r2.Err == nil {
		t.Fatal("shouldn't have returned a member")
	}
}

func TestGetMemberCount(t *testing.T) {
	Setup()

	teamId := model.NewId()

	c1 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Channel1",
		Name:        "a" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	Must(store.Channel().Save(&c1))

	c2 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Channel2",
		Name:        "a" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	Must(store.Channel().Save(&c2))

	u1 := &model.User{
		Email:    model.NewId(),
		DeleteAt: 0,
	}
	Must(store.User().Save(u1))
	Must(store.Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u1.Id}))

	m1 := model.ChannelMember{
		ChannelId:   c1.Id,
		UserId:      u1.Id,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	Must(store.Channel().SaveMember(&m1))

	if result := <-store.Channel().GetMemberCount(c1.Id, false); result.Err != nil {
		t.Fatalf("failed to get member count: %v", result.Err)
	} else if result.Data.(int64) != 1 {
		t.Fatalf("got incorrect member count %v", result.Data)
	}

	u2 := model.User{
		Email:    model.NewId(),
		DeleteAt: 0,
	}
	Must(store.User().Save(&u2))
	Must(store.Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u2.Id}))

	m2 := model.ChannelMember{
		ChannelId:   c1.Id,
		UserId:      u2.Id,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	Must(store.Channel().SaveMember(&m2))

	if result := <-store.Channel().GetMemberCount(c1.Id, false); result.Err != nil {
		t.Fatalf("failed to get member count: %v", result.Err)
	} else if result.Data.(int64) != 2 {
		t.Fatalf("got incorrect member count %v", result.Data)
	}

	// make sure members of other channels aren't counted
	u3 := model.User{
		Email:    model.NewId(),
		DeleteAt: 0,
	}
	Must(store.User().Save(&u3))
	Must(store.Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u3.Id}))

	m3 := model.ChannelMember{
		ChannelId:   c2.Id,
		UserId:      u3.Id,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	Must(store.Channel().SaveMember(&m3))

	if result := <-store.Channel().GetMemberCount(c1.Id, false); result.Err != nil {
		t.Fatalf("failed to get member count: %v", result.Err)
	} else if result.Data.(int64) != 2 {
		t.Fatalf("got incorrect member count %v", result.Data)
	}

	// make sure inactive users aren't counted
	u4 := &model.User{
		Email:    model.NewId(),
		DeleteAt: 10000,
	}
	Must(store.User().Save(u4))
	Must(store.Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u4.Id}))

	m4 := model.ChannelMember{
		ChannelId:   c1.Id,
		UserId:      u4.Id,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	Must(store.Channel().SaveMember(&m4))

	if result := <-store.Channel().GetMemberCount(c1.Id, false); result.Err != nil {
		t.Fatalf("failed to get member count: %v", result.Err)
	} else if result.Data.(int64) != 2 {
		t.Fatalf("got incorrect member count %v", result.Data)
	}
}

func TestUpdateExtrasByUser(t *testing.T) {
	Setup()

	teamId := model.NewId()

	c1 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Channel1",
		Name:        "a" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	Must(store.Channel().Save(&c1))

	c2 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Channel2",
		Name:        "a" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	Must(store.Channel().Save(&c2))

	u1 := &model.User{
		Email:    model.NewId(),
		DeleteAt: 0,
	}
	Must(store.User().Save(u1))
	Must(store.Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u1.Id}))

	m1 := model.ChannelMember{
		ChannelId:   c1.Id,
		UserId:      u1.Id,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	Must(store.Channel().SaveMember(&m1))

	u1.DeleteAt = model.GetMillis()
	Must(store.User().Update(u1, true))

	if result := <-store.Channel().ExtraUpdateByUser(u1.Id, u1.DeleteAt); result.Err != nil {
		t.Fatalf("failed to update extras by user: %v", result.Err)
	}

	u1.DeleteAt = 0
	Must(store.User().Update(u1, true))

	if result := <-store.Channel().ExtraUpdateByUser(u1.Id, u1.DeleteAt); result.Err != nil {
		t.Fatalf("failed to update extras by user: %v", result.Err)
	}
}

func TestChannelStoreSearchMore(t *testing.T) {
	Setup()

	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "ChannelA"
	o1.Name = "a" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o1))

	o2 := model.Channel{}
	o2.TeamId = model.NewId()
	o2.DisplayName = "Channel2"
	o2.Name = "a" + model.NewId() + "b"
	o2.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o2))

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = model.NewId()
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&m1))

	m2 := model.ChannelMember{}
	m2.ChannelId = o1.Id
	m2.UserId = model.NewId()
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&m2))

	m3 := model.ChannelMember{}
	m3.ChannelId = o2.Id
	m3.UserId = model.NewId()
	m3.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&m3))

	o3 := model.Channel{}
	o3.TeamId = o1.TeamId
	o3.DisplayName = "ChannelA"
	o3.Name = "a" + model.NewId() + "b"
	o3.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o3))

	o4 := model.Channel{}
	o4.TeamId = o1.TeamId
	o4.DisplayName = "ChannelB"
	o4.Name = "a" + model.NewId() + "b"
	o4.Type = model.CHANNEL_PRIVATE
	Must(store.Channel().Save(&o4))

	o5 := model.Channel{}
	o5.TeamId = o1.TeamId
	o5.DisplayName = "ChannelC"
	o5.Name = "a" + model.NewId() + "b"
	o5.Type = model.CHANNEL_PRIVATE
	Must(store.Channel().Save(&o5))

	if result := <-store.Channel().SearchMore(m1.UserId, o1.TeamId, "ChannelA"); result.Err != nil {
		t.Fatal(result.Err)
	} else {
		channels := result.Data.(*model.ChannelList)
		if len(*channels) == 0 {
			t.Fatal("should not be empty")
		}

		if (*channels)[0].Name != o3.Name {
			t.Fatal("wrong channel returned")
		}
	}

	if result := <-store.Channel().SearchMore(m1.UserId, o1.TeamId, o4.Name); result.Err != nil {
		t.Fatal(result.Err)
	} else {
		channels := result.Data.(*model.ChannelList)
		if len(*channels) != 0 {
			t.Fatal("should be empty")
		}
	}

	if result := <-store.Channel().SearchMore(m1.UserId, o1.TeamId, o3.Name); result.Err != nil {
		t.Fatal(result.Err)
	} else {
		channels := result.Data.(*model.ChannelList)
		if len(*channels) == 0 {
			t.Fatal("should not be empty")
		}

		if (*channels)[0].Name != o3.Name {
			t.Fatal("wrong channel returned")
		}
	}

}

func TestChannelStoreSearchInTeam(t *testing.T) {
	Setup()

	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "ChannelA"
	o1.Name = "a" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o1))

	o2 := model.Channel{}
	o2.TeamId = model.NewId()
	o2.DisplayName = "Channel2"
	o2.Name = "a" + model.NewId() + "b"
	o2.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o2))

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = model.NewId()
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&m1))

	m2 := model.ChannelMember{}
	m2.ChannelId = o1.Id
	m2.UserId = model.NewId()
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&m2))

	m3 := model.ChannelMember{}
	m3.ChannelId = o2.Id
	m3.UserId = model.NewId()
	m3.NotifyProps = model.GetDefaultChannelNotifyProps()
	Must(store.Channel().SaveMember(&m3))

	o3 := model.Channel{}
	o3.TeamId = o1.TeamId
	o3.DisplayName = "ChannelA"
	o3.Name = "a" + model.NewId() + "b"
	o3.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o3))

	o4 := model.Channel{}
	o4.TeamId = o1.TeamId
	o4.DisplayName = "ChannelB"
	o4.Name = "a" + model.NewId() + "b"
	o4.Type = model.CHANNEL_PRIVATE
	Must(store.Channel().Save(&o4))

	o5 := model.Channel{}
	o5.TeamId = o1.TeamId
	o5.DisplayName = "ChannelC"
	o5.Name = "a" + model.NewId() + "b"
	o5.Type = model.CHANNEL_PRIVATE
	Must(store.Channel().Save(&o5))

	if result := <-store.Channel().SearchInTeam(o1.TeamId, "ChannelA"); result.Err != nil {
		t.Fatal(result.Err)
	} else {
		channels := result.Data.(*model.ChannelList)
		if len(*channels) != 2 {
			t.Fatal("wrong length")
		}
	}

	if result := <-store.Channel().SearchInTeam(o1.TeamId, ""); result.Err != nil {
		t.Fatal(result.Err)
	} else {
		channels := result.Data.(*model.ChannelList)
		if len(*channels) == 0 {
			t.Fatal("should not be empty")
		}
	}

	if result := <-store.Channel().SearchInTeam(o1.TeamId, "blargh"); result.Err != nil {
		t.Fatal(result.Err)
	} else {
		channels := result.Data.(*model.ChannelList)
		if len(*channels) != 0 {
			t.Fatal("should be empty")
		}
	}
}

func TestChannelStoreGetMembersByIds(t *testing.T) {
	Setup()

	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "ChannelA"
	o1.Name = "a" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	Must(store.Channel().Save(&o1))

	m1 := &model.ChannelMember{ChannelId: o1.Id, UserId: model.NewId(), NotifyProps: model.GetDefaultChannelNotifyProps()}
	Must(store.Channel().SaveMember(m1))

	if r := <-store.Channel().GetMembersByIds(m1.ChannelId, []string{m1.UserId}); r.Err != nil {
		t.Fatal(r.Err)
	} else {
		rm1 := (*r.Data.(*model.ChannelMembers))[0]

		if rm1.ChannelId != m1.ChannelId {
			t.Fatal("bad team id")
		}

		if rm1.UserId != m1.UserId {
			t.Fatal("bad user id")
		}
	}

	m2 := &model.ChannelMember{ChannelId: o1.Id, UserId: model.NewId(), NotifyProps: model.GetDefaultChannelNotifyProps()}
	Must(store.Channel().SaveMember(m2))

	if r := <-store.Channel().GetMembersByIds(m1.ChannelId, []string{m1.UserId, m2.UserId, model.NewId()}); r.Err != nil {
		t.Fatal(r.Err)
	} else {
		rm := (*r.Data.(*model.ChannelMembers))

		if len(rm) != 2 {
			t.Fatal("return wrong number of results")
		}
	}

	if r := <-store.Channel().GetMembersByIds(m1.ChannelId, []string{}); r.Err == nil {
		t.Fatal("empty user ids - should have failed")
	}
}
