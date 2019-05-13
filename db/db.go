package db

import (
	"errors"
	"db2rest/conf"
	"strings"
	"log"
	"time"
	"regexp"
	"net/url"
	"database/sql"
)

type Client struct {
	conf		*conf.Conf
	regex1		*regexp.Regexp
	regex2		*regexp.Regexp
	db *sql.DB
}

func New(conf *conf.Conf) (*Client, error) {
	c := &Client{conf: conf}
	dsn := conf.GetString("db.url", "")
	if dsn == "" {
		return nil, errors.New("db.url is not set")
	}
	url, err := url.Parse(dsn)
	if err != nil {
		return nil, errors.New("db.url in wrong format")
	}
	log.Printf("open db %s://%s:****@%s%s\n", url.Scheme, url.User.Username(), url.Host, url.RequestURI())
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(conf.GetInt("db.max_open_conn", 3))
	db.SetMaxIdleConns(conf.GetInt("db.max_idle_conn", 3))
	db.SetConnMaxLifetime(time.Duration(conf.GetInt("db.max_lifetime_minute", 5))*time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	c.db = db
	c.regex1 = regexp.MustCompile(`\s*\r?\n`)
	c.regex2 = regexp.MustCompile(`\s*\r?\n\s*`)

	return c, nil
}

func newRow(n int) ([]*[]byte, []interface{}) {
	vals1 := make([]*[]byte, n)
	vals2 := make([]interface{}, n)
	for i := 0; i < n; i++ {
		r := new([]byte)
		vals1[i] = r
		vals2[i] = r
	}
	return vals1, vals2
}

func resetRow(row []*[]byte) {
	for _, val := range row {
		(*val) = (*val)[:0]
	}
}

func (c *Client) Query(out Output) {
	sql, err := out.SQL()
	if err != nil {
		out.Error(err)
		return
	}
	log.Printf("sql: %s\n", sql)

	rows, err := c.db.Query(sql)
	if err != nil {
		out.Error(err)
		return
	}
	defer rows.Close()

	cols, err := rows.ColumnTypes()
	if err != nil {
		out.Error(err)
		return
	}
	if err := out.Columns(cols); err != nil {
		out.Error(err)
		return
	}

	val1, val2 := newRow(len(cols))
	for rows.Next() {
		resetRow(val1)
		if err := rows.Scan(val2...); err != nil {
			out.Error(err)
			return
		}
		if err := out.Row(val1); err != nil {
			out.Error(err)
			return
		}
	}
	if err := rows.Err(); err != nil {
		out.Error(err)
		return
	}

	out.End()
}

func (c *Client) Exec(out Output) {
	sql, err := out.SQL()
	if err != nil {
		out.Error(err)
		return
	}
	log.Printf("sql: %s\n", sql)

	res, err := c.db.Exec(sql)
	if err != nil {
		out.Error(err)
		return
	}

	// lastId, err := res.LastInsertId()
	// if err != nil {
	// 	out.Error(err)
	// 	return
	// }
	rowCnt, err := res.RowsAffected()
	if err != nil {
		out.Error(err)
		return
	}

	out.Affected(0, rowCnt)
	out.End()
}

func (c *Client) Close() (err error) {
	if c.db != nil {
		log.Println("closing db")
		err = c.db.Close();
	}
	return 
}

func (c *Client) FormatSQL(sql string) string {
	if c.conf.GetBool("db.replace_newline_with_space", false) {
		sql = c.regex2.ReplaceAllString(sql, " ")
		sql = strings.TrimSpace(sql)
	} else if c.conf.GetBool("db.remove_empty_line", true) {
		sql = c.regex1.ReplaceAllString(sql, "\n")
	}
	return sql
}

type Output interface {
	SQL() (string, error)

	Columns([]*sql.ColumnType) error
	Row([]*[]byte) error

	Affected(lastInsertId, rowsAffected int64)

	Error(error)
	End()
}
