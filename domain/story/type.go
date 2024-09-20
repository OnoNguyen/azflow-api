package story

type DmAudio struct {
	Url           string
	TranscriptUrl string
	Title         string
	Id            int
}

type DbAudio struct {
	Id       int    `db:"id"`
	ExtId    string `db:"ext_id"`
	Title    string `db:"title"`
	FileName string `db:"file_name"`
}
