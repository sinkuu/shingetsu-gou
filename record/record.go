/*
 * Copyright (c) 2015, Shinya Yagyu
 * All rights reserved.
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice,
 *    this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 *    this list of conditions and the following disclaimer in the documentation
 *    and/or other materials provided with the distribution.
 * 3. Neither the name of the copyright holder nor the names of its
 *    contributors may be used to endorse or promote products derived from this
 *    software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */

package record

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/shingetsu-gou/shingetsu-gou/cfg"
	"github.com/shingetsu-gou/shingetsu-gou/db"
	"github.com/shingetsu-gou/shingetsu-gou/node"
	"github.com/shingetsu-gou/shingetsu-gou/util"
)

var cachedRule = util.NewRegexpList(cfg.SpamList)

//Record represents one record.
type Record struct {
	*Head
	contents map[string]string
	keyOrder []string
}

//NewIDstr parse idstr unixtime+"_"+md5(bodystr)), set stamp and id, and return record obj.
//if parse failes returns nil.
func NewIDstr(datfile, idstr string) (*Record, error) {
	if idstr == "" {
		return New(datfile, "", 0), nil
	}
	buf := strings.Split(idstr, "_")
	if len(buf) != 2 {
		err := errors.New("bad format")
		log.Println(idstr, ":bad format")
		return nil, err
	}
	var stamp int64
	var err error
	if stamp, err = strconv.ParseInt(buf[0], 10, 64); err != nil {
		log.Println(idstr, ":bad format")
		return nil, err
	}
	return New(datfile, buf[1], stamp), nil
}

//New makes Record struct.
func New(datfile, id string, stamp int64) *Record {
	return &Record{
		Head: &Head{
			Datfile: datfile,
			Stamp:   stamp,
			ID:      id,
		},
	}
}

//CopyHead copies and  returns Head.
func (r *Record) CopyHead() Head {
	return Head{
		Datfile: r.Datfile,
		Stamp:   r.Stamp,
		ID:      r.ID,
	}
}

//Make makes and returns record from Recstr
func Make(line string) *Record {
	line = strings.TrimRight(line, "\r\n")
	buf := strings.Split(line, "<>")
	if len(buf) <= 2 || buf[0] == "" || buf[1] == "" || buf[2] == "" {
		return nil
	}
	idstr := buf[0] + "_" + buf[1]
	dec := util.FileDecode(buf[2])
	if dec == "" || !strings.HasPrefix(buf[2], "thread_") {
		//		log.Println("illegal format",buf[2])
		return nil
	}
	buf[2] = util.FileEncode("thread", dec)
	vr, err := NewIDstr(buf[2], idstr)
	if err != nil {
		return nil
	}
	if err := vr.Parse(line); err != nil {
		log.Println(err)
		return nil
	}
	return vr
}

//bodystr returns body part of one line in the record file.
func (r *Record) bodystr() string {
	rs := make([]string, len(r.contents))
	for i, k := range r.keyOrder {
		rs[i] = k + ":" + r.contents[k]
	}
	return strings.Join(rs, "<>")
}

//HasBodyValue returns true if key k exists
//used in templates
func (r *Record) HasBodyValue(k string) bool {
	if _, exist := r.contents[k]; exist {
		return true
	}
	return false
}

//GetBodyValue returns value of key k
//return def if not exists.
func (r *Record) GetBodyValue(k string, def string) string {
	if v, exist := r.contents[k]; exist {
		return v
	}
	return def
}

//Recstr returns one line in the record file.
func (r *Record) Recstr() string {
	return fmt.Sprintf("%d<>%s<>%s", r.Stamp, r.ID, r.bodystr())
}

//Parse parses one line in record file and response of /recent/ and set params to record r.
func (r *Record) Parse(recstr string) error {
	var err error
	recstr = strings.TrimRight(recstr, "\r\n")
	tmp := strings.Split(recstr, "<>")
	if len(tmp) < 2 {
		errr := errors.New(recstr + ":bad format")
		log.Println(errr)
		return errr
	}
	stamp, err := strconv.ParseInt(tmp[0], 10, 64)
	if err != nil {
		log.Println(tmp[0], "bad format")
		return err
	}
	if r.Stamp == 0 {
		r.Stamp = stamp
	}
	if r.Stamp != stamp {
		log.Println("stamp unmatch")
		return errors.New("stamp unmatch")
	}
	if r.ID == "" {
		r.ID = tmp[1]
	}
	if r.ID != tmp[1] {
		log.Println("ID unmatch")
		return errors.New("stamp unmatch")
	}
	r.contents = make(map[string]string)
	r.keyOrder = nil
	//reposense of recentlist  : stamp<>id<>thread_***<>tag:***
	//record str : stamp<>id<>body:***<>...
	for _, kv := range tmp[2:] {
		buf := strings.SplitN(kv, ":", 2)
		if len(buf) < 2 {
			continue
		}
		buf[1] = strings.Replace(buf[1], "<br>", "\n", -1)
		buf[1] = strings.Replace(buf[1], "<", "&lt;", -1)
		buf[1] = strings.Replace(buf[1], ">", "&gt;", -1)
		buf[1] = strings.Replace(buf[1], "\n", "<br>", -1)
		if util.HasString(r.keyOrder, buf[0]) {
			err := errors.New("duplicate keys")
			log.Println(err)
			return err
		}
		r.keyOrder = append(r.keyOrder, buf[0])
		r.contents[buf[0]] = buf[1]
	}
	return nil
}

//Load loads a record file and parses it.
func (r *Record) Load() error {
	if !r.Exists() {
		err := r.Remove()
		if err != nil {
			log.Println(err)
		}
		return errors.New("file not found")
	}
	body, err := db.String("select Body from record where Thread=? and Hash=? and Stamp=?", r.Datfile, r.ID, r.Stamp)
	if err != nil {
		log.Println(err)
		return err
	}
	return r.Parse(fmt.Sprintf("%d<>%s<>%s", r.Stamp, r.ID, body))
}

//ShortPubkey returns short version of pubkey.
func (r *Record) ShortPubkey() string {
	if v, exist := r.contents["pubkey"]; exist {
		return util.CutKey(v)
	}
	return ""
}

//Build sets params in record from args and return id.
func (r *Record) Build(stamp int64, body map[string]string, passwd string) string {
	r.contents = make(map[string]string)
	r.Stamp = stamp
	for key, value := range body {
		if value == "" {
			continue
		}
		r.contents[key] = value
		r.keyOrder = append(r.keyOrder, key)
	}
	if passwd != "" {
		k := util.MakePrivateKey(passwd)
		pubkey, _ := k.GetKeys()
		md := util.MD5digest(r.bodystr())
		sign := k.Sign(md)
		r.contents["pubkey"] = pubkey
		r.contents["sign"] = sign
		r.contents["target"] = strings.Join(r.keyOrder, ",")
		r.keyOrder = append(r.keyOrder, "pubkey")
		r.keyOrder = append(r.keyOrder, "sign")
		r.keyOrder = append(r.keyOrder, "target")
	}

	id := util.MD5digest(r.bodystr())
	r.ID = id
	return r.ID
}

//md5check return true if md5 of bodystr is same as r.id.
func (r *Record) md5check() bool {
	return util.MD5digest(r.bodystr()) == r.ID
}

//AttachPath returns attach path
//by creating path from args.
func (r *Record) AttachPath(thumbnailSize string) string {
	suffix := r.GetBodyValue("suffix", "")
	if suffix == "" {
		return ""
	}
	reg := regexp.MustCompile(`[^-_.A-Za-z0-9]`)
	reg.ReplaceAllString(suffix, "")
	if thumbnailSize != "" {
		return "s" + r.Idstr() + "." + thumbnailSize + "." + suffix
	}
	return r.Idstr() + "." + suffix
}

//Sync saves Recstr to the file. if attached file exists, saves it to attached path.
//if signed, also saves body part.
func (r *Record) Sync() {
	cnt, err := db.Int64("select count(*) from record where Thread=? and Stamp=? and Hash=?",
		r.Datfile, r.Stamp, r.ID)
	if err != nil {
		log.Print(err)
	}
	if cnt > 0 {
		return
	}
	_, err = db.DB.Exec("insert into record(Stamp,Hash,Thread,Body,Deleted) values(?,?,?,?,0)", r.Stamp, r.ID, r.Datfile, r.bodystr())
	if err != nil {
		log.Print(err)
	}
}

//Getbody retuns contents of rec after loading if needed.
func (r *Record) Getbody() string {
	err := r.Load()
	if err != nil {
		log.Println(err)
	}
	return r.Recstr()
}

//checkSign check signature in the record is valid.
func (r *Record) checkSign() bool {

	for _, k := range []string{"pubkey", "sign", "target"} {
		if _, exist := r.contents[k]; !exist {
			return false
		}
	}
	ts := strings.Split(r.contents["target"], ",")
	targets := make([]string, len(ts))
	for i, t := range ts {
		if _, exist := r.contents[t]; !exist {
			return false
		}
		targets[i] = t + ":" + r.contents[t]
	}
	md := util.MD5digest(strings.Join(targets, "<>"))
	return util.Verify(md, r.contents["sign"], r.contents["pubkey"])
}

//meets checks the record meets conditions of args
func (r *Record) meets(begin, end int64) bool {
	md5ok := r.md5check()
	if begin > r.Stamp || (end > 0 && r.Stamp > end) {
		log.Println("stamp range NG", begin, end, r.Stamp)
		return false
	}
	if !md5ok {
		log.Println("md5 NG")
		return false
	}
	return true
}

//IsSpam returns true if Recstr is listed in spam.txt
func (r *Record) IsSpam() bool {
	return cachedRule.Check(r.Recstr())
}

//MakeAttachLink makes and returns attached file link.
func (r *Record) MakeAttachLink(sakuHost string) string {
	if r.GetBodyValue("attach", "") == "" {
		return ""
	}
	url := fmt.Sprintf("http://%s/thread.cgi/%s/%s/%d.%s",
		sakuHost, r.Datfile, r.ID, r.Stamp, r.GetBodyValue("suffix", cfg.SuffixTXT))
	return "<br><br>[Attached]<br>" + url
}

//BodyString retuns bodystr not including attach field, and shorten pubkey.
func (r *Record) BodyString() string {
	buf := []string{
		strconv.FormatInt(r.Stamp, 10),
		r.ID,
	}
	for _, k := range r.keyOrder {
		switch k {
		case "attach":
			buf = append(buf, "attach:1")
		case "sign":
		case "pubkey":
			if r.checkSign() {
				shortKey := util.CutKey(r.contents["pubkey"])
				buf = append(buf, "pubkey:"+shortKey)
			}
		default:
			buf = append(buf, k+":"+r.contents[k])
		}
	}
	return strings.Join(buf, "<>")
}

//GetData gets records from node n and checks its is same as stamp and id in args.
//save recs if success. returns errSpam or errGet.
func (r *Record) GetData(n *node.Node) error {
	res, err := n.Talk(fmt.Sprintf("/get/%s/%d/%s", r.Datfile, r.Stamp, r.ID), nil)
	if len(res) == 0 {
		err = errors.New("no response")
	}
	if err != nil {
		log.Println(err)
		return cfg.ErrGet
	}
	if err = r.Parse(res[0]); err != nil {
		return cfg.ErrGet
	}
	r.Sync()
	return r.CheckData(-1, -1)
}

//CheckData makes records from res and checks its records meets condisions of args.
//adds the rec to cache if meets conditions.
//if spam or big data, remove the rec from disk.
//returns count of added records to the cache and spam/getting error.
func (r *Record) CheckData(begin, end int64) error {
	if !r.meets(begin, end) {
		return cfg.ErrGet
	}
	if len(r.Recstr()) > cfg.RecordLimit<<10 || r.IsSpam() {
		log.Printf("warning:%s/%s:too large or spam record", r.Datfile, r.Idstr())
		errr := r.Remove()
		if errr != nil {
			log.Println(errr)
		}
		return cfg.ErrSpam
	}
	return nil
}

//InRange returns true if stamp  is in begin~end and idstr has id.
func (r *Record) InRange(begin, end int64, id string) bool {
	return begin <= r.Stamp && r.Stamp <= end && (id == "" || strings.HasSuffix(r.Idstr(), id))
}
