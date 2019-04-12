package loader

import (
	"database/sql"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	check "github.com/pingcap/check"
)

type LoadSuite struct {
}

var _ = check.Suite(&LoadSuite{})

func (cs *LoadSuite) SetUpTest(c *check.C) {
}

func (cs *LoadSuite) TearDownTest(c *check.C) {
}

func (cs *LoadSuite) TestNewClose(c *check.C) {
	db, _, err := sqlmock.New()
	c.Assert(err, check.IsNil)

	loader, err := NewLoader(db)
	c.Assert(err, check.IsNil)

	loader.Close()
}

type groupDMLsSuite struct{}

var _ = check.Suite(&groupDMLsSuite{})

func (s *groupDMLsSuite) TestSingleDMLsOnlyIfDisableMerge(c *check.C) {
	ld := Loader{merge: false}
	dmls := []*DML{
		&DML{Tp: UpdateDMLType},
		&DML{Tp: UpdateDMLType},
		&DML{Tp: InsertDMLType},
	}
	batch, single := ld.groupDMLs(dmls)
	c.Assert(batch, check.HasLen, 0)
	c.Assert(single, check.HasLen, 3)
}

func (s *groupDMLsSuite) TestGroupByTableName(c *check.C) {
	ld := Loader{merge: true}
	canBatch := tableInfo{primaryKey: &indexInfo{}}
	onlySingle := tableInfo{}
	dmls := []*DML{
		&DML{Table: "test1", info: &canBatch},
		&DML{Table: "test1", info: &canBatch},
		&DML{Table: "test2", info: &onlySingle},
		&DML{Table: "test1", info: &canBatch},
		&DML{Table: "test2", info: &onlySingle},
	}
	batch, single := ld.groupDMLs(dmls)
	c.Assert(batch, check.HasLen, 1)
	c.Assert(batch[dmls[0].TableName()], check.HasLen, 3)
	c.Assert(single, check.HasLen, 2)
}

type getTblInfoSuite struct{}

var _ = check.Suite(&getTblInfoSuite{})

func (s *getTblInfoSuite) TestShouldCacheResult(c *check.C) {
	origGet := utilGetTableInfo
	nCalled := 0
	utilGetTableInfo = func(db *sql.DB, schema string, table string) (info *tableInfo, err error) {
		nCalled++
		return &tableInfo{columns: []string{"id", "name"}}, nil
	}
	defer func() {
		utilGetTableInfo = origGet
	}()
	ld := Loader{}

	info, err := ld.getTableInfo("test", "contacts")
	c.Assert(err, check.IsNil)
	c.Assert(info.columns[1], check.Equals, "name")

	info, err = ld.getTableInfo("test", "contacts")
	c.Assert(info.columns[1], check.Equals, "name")

	c.Assert(nCalled, check.Equals, 1)
}
