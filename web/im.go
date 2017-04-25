package web

import (
	"database/sql"
	"fmt"
	"strconv"
)

type Root struct {
	Code int             `json:"code"`
	Msg  string           `json:"msg"`
	Data interface{}     `json:"data"`
}

type ListType struct {
	Mine   map[string]string      `json:"mine"`
	Friend []FriendType            `json:"friend"`
	Group  []map[string]string     `json:"group"`
}

type FriendType struct {
	Groupname string               `json:"groupname"`
	Online    int			`json:"online"`
	Id        int			`json:"id"`
	List      []map[string]string	`json:"list"`
}

type MemberType struct {
	Owner     map[string]string    `json:"owner"`
	Members    int			`json:"members"`
	List      []map[string]string	`json:"list"`
}


func GetUserRow(db *sql.DB, sid string) map[string]string {

	sql_str := `select id,nick,status ,sign, avatar,token  from user where sid=?`
	var id, nick, status, sign, avatar,token string
	record := make(map[string]string)
	err := db.QueryRow(sql_str, sid).Scan(&id, &nick, &status, &sign, &avatar,&token )
	if err!=nil {
		fmt.Println("getUserRow err:", err.Error())
		return record
	}
	record["id"] = id
	record["username"] = nick
	record["sign"] = sign
	record["status"] = status
	record["sid"] = sid
	record["avatar"] = avatar
	record["token"] = token

	return record
}

func getMyContacts(db *sql.DB, uid int) []map[string]string {

	sql_str := "SELECT  u.id,u.nick as nick,u.avatar,u.sign,c.group_id,u.sid  FROM `contacts` c LEFT JOIN `user` u on u.id =c.uid WHERE  c.master_uid=?"

	contact_records := make([]map[string]string, 0)
	rows, err := db.Query(sql_str, uid)
	if err != nil {

		return contact_records
	}
	for rows.Next() {
		//将行数据保存到record字典
		var id, nick, avatar, sign, group_id,sid string
		record := make(map[string]string)
		rows.Scan(&id, &nick, &avatar, &sign, &group_id, &sid)

		record["id"] = id
		record["username"] = nick
		record["avatar"] = avatar
		record["sign"] = sign
		record["group_id"] = group_id
		record["sid"] = sid
		contact_records = append(contact_records, record)

	}
	return contact_records

}

func getMyGroup(db *sql.DB, uid int) []map[string]string {

	sql_str := "SELECT  id,title as groupname  FROM `contact_group` WHERE uid=? "
	my_group_records := make([]map[string]string, 0)
	rows, err := db.Query(sql_str, uid)
	if err != nil {
		return my_group_records
	}
	for rows.Next() {
		//将行数据保存到record字典
		var gid, groupname string
		record := make(map[string]string)
		err = rows.Scan(&gid, &groupname)
		if err != nil {
			fmt.Println( "服务器错误@"+err.Error())
			return my_group_records
		}
		record["id"] = gid
		record["groupname"] = groupname
		fmt.Println(record)
		my_group_records = append(my_group_records, record)
	}
	return my_group_records
}

func getFriends(db *sql.DB, uid int) []FriendType {

	friends := make([]FriendType,0)

	// 获取所属的联系人列表（未分组）
	contact_records := getMyContacts(db, uid)

	// 获取分组
	my_group_records := getMyGroup(db, uid)
	var friend 	FriendType
	for _, group := range my_group_records {
		friend = FriendType{}
		friend.Groupname = group[`groupname`]
		friend.Id ,_= strconv.Atoi(group[`id`])
		friend.Online = 1
		tmp_list := make([]map[string]string, 0)

		for _ ,c := range contact_records {
			group_id,_ :=strconv.Atoi(c[`group_id`])
			if group_id == friend.Id {
				tmp_list = append(tmp_list, c)
				//contact_records = append(contact_records[:_k], contact_records[_k+1:]...)
			}
		}
		friend.List = tmp_list
		friends = append(friends, friend)
	}

	return friends
}

func getMyGroups(db *sql.DB, uid int) []map[string]string {

	sql_str := "SELECT id,channel_id,pic as avatar,title  FROM `global_group` WHERE  id in( SELECT `group_id` FROM `user_join_group` WHERE `uid`=? )"
	join_group_records := make([]map[string]string, 0)
	rows, err := db.Query(sql_str, uid)
	if err != nil {
		fmt.Println( 504, "服务器错误@"+err.Error())
		return join_group_records
	}
	for rows.Next() {
		//将行数据保存到record字典
		var cid, channel_id, avatar,title string
		record := make(map[string]string)
		err = rows.Scan(&cid, &channel_id, &avatar,&title)
		if err != nil {
			fmt.Println( 505, "服务器错误@"+err.Error())
			return join_group_records
		}
		record["id"] = cid
		record["channel_id"] = channel_id
		record["avatar"] = avatar
		record["groupname"] =  title
		//fmt.Println(record)
		join_group_records = append(join_group_records, record)
	}
	fmt.Println(join_group_records)
	return join_group_records

}

func getMembers( db  *sql.DB, member_id int) []map[string]string {

	sql_str := "SELECT U.id,U.nick, U.sign, U.avatar,U.sid  FROM `user_join_group` G LEFT JOIN user U on G.uid=U.id WHERE  group_id=?"
	members := make([]map[string]string, 0)
	rows, err := db.Query(sql_str, member_id)
	if err != nil {
		fmt.Println( 504, "服务器错误@"+err.Error())
		return members
	}
	for rows.Next() {
		//将行数据保存到record字典
		var id, nick,sign, avatar ,sid string
		record := make(map[string]string)
		err = rows.Scan(&id, &nick, &sign,&avatar,&sid)
		if err != nil {
			fmt.Println( 505, "服务器错误@"+err.Error())
			return members
		}
		record["id"] = id
		record["sign"] = sign
		record["avatar"] = avatar
		record["username"] =  nick
		//fmt.Println(record)
		members = append(members, record)
	}
	fmt.Println(members)
	return members

}