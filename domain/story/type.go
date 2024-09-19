package story

type DmAudio struct {
	Url        string
	CaptionUrl string
	Title      string
	Id         int
}

type PgAudio struct {
	Id       int    `db:"id"`
	ExtId    string `db:"ext_id"`
	Title    string `db:"title"`
	FileName string `db:"file_name"`
}
