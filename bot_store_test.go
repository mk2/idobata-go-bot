package idobot_test

import (
	"os"
	"testing"

	"github.com/mk2/idobot"
)

func TestIdobot_DBを作成する(t *testing.T) {
	_, err := idobot.NewStore("./test.db")
	defer os.Remove("./test.db")
	if err != nil {
		t.Fatalf("dbの作成に失敗しました。\n")
	}
}

func TestIdobot_DBに書き込む(t *testing.T) {
	st, err := idobot.NewStore("./test.db")
	defer os.Remove("./test.db")
	if err != nil {
		t.Fatalf("dbの作成に失敗しました。\n")
	}

	err = st.Save("name", "hoge")
	if err != nil {
		t.Errorf("dbの書き込みに失敗しました。\n")
	}
}

func TestIdobot_DBから読み込む(t *testing.T) {
	st, err := idobot.NewStore("./test.db")
	defer os.Remove("./test.db")
	if err != nil {
		t.Fatalf("dbの作成に失敗しました。\n")
	}
	err = st.Save("name", "hoge")
	if err != nil {
		t.Errorf("dbの書き込みに失敗しました。\n")
	}

	content, err := st.Read("name")
	if err != nil {
		t.Errorf("dbの読み込みに失敗しました。\n")
	}

	if content != "hoge" {
		t.Errorf("dbの読み込みに失敗しました。\n")
	}
}
