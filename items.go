package main

func (p *Player) EquipItem(item Item, chatID int64) {
	p.Inventory = append(p.Inventory, item)
	switch item.Stat {
	case "attack":
		p.Attack += item.Value
	case "defense":
		p.Defense += item.Value
	case "healplayer":
		p.MaxHP += item.Value
	}
}

