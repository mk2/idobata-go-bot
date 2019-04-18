package idobot

/*
botの情報を返す機能を集めたファイル
*/

func (bot *botImpl) RoomIDs() []int {
	keys := make([]int, 0, len(bot.roomIDs.set))
	for k := range bot.roomIDs.set {
		keys = append(keys, k)
	}
	return keys
}

func (bot *botImpl) BotID() int {
	return bot.botID
}

func (bot *botImpl) BotName() string {
	return bot.botName
}
