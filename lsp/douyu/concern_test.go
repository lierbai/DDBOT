package douyu

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const testRoomStr = "9617408"

func TestConcern(t *testing.T) {
	logrus.SetLevel(logrus.TraceLevel)
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	testEventChan := make(chan concern.Event, 16)
	testNotifyChan := make(chan concern.Notify)

	c := NewConcern(testNotifyChan)

	assert.NotNil(t, c.GetStateManager())

	_testRoom, err := c.ParseId(testRoomStr)
	assert.Nil(t, err)
	testRoom := _testRoom.(int64)

	c.StateManager.UseNotifyGenerator(c.notifyGenerator())
	c.StateManager.UseFreshFunc(func(eventChan chan<- concern.Event) {
		for e := range testEventChan {
			eventChan <- e
		}
	})

	assert.Nil(t, c.StateManager.Start())
	defer c.Stop()
	defer close(testEventChan)

	_, err = c.Add(nil, test.G1, testRoom, Live)
	assert.Nil(t, err)

	liveInfo, err := c.FindOrLoadRoom(testRoom)
	assert.Nil(t, err)
	assert.NotNil(t, liveInfo)
	assert.Equal(t, testRoom, liveInfo.RoomId)
	assert.Equal(t, "斗鱼官方视频号", liveInfo.RoomName)

	identityInfo, err := c.Get(testRoom)
	assert.Nil(t, err)
	assert.EqualValues(t, liveInfo.GetRoomId(), identityInfo.GetUid())
	assert.EqualValues(t, liveInfo.GetNickname(), identityInfo.GetName())

	identityInfos, ctypes, err := c.List(test.G1, Live)
	assert.Nil(t, err)
	assert.Len(t, identityInfos, 1)
	assert.Len(t, ctypes, 1)
	assert.Equal(t, Live, ctypes[0])

	info := identityInfos[0]
	assert.Equal(t, testRoom, info.GetUid())
	assert.Equal(t, "斗鱼官方视频号", info.GetName())

	liveInfo.ShowStatus = ShowStatus_Living
	liveInfo.VideoLoop = VideoLoopStatus_Off

	testEventChan <- liveInfo

	select {
	case notify := <-testNotifyChan:
		assert.Equal(t, test.G1, notify.GetGroupCode())
	case <-time.After(time.Second):
		assert.Fail(t, "no notify received")
	}

	identityInfo, err = c.Remove(nil, test.G1, testRoom, Live)
	assert.Nil(t, err)
	assert.EqualValues(t, testRoom, identityInfo.GetUid())

	identityInfo, err = c.Remove(nil, test.G1, testRoom, Live)
	assert.NotNil(t, err)
}
