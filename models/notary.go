package models

import (
	"github.com/filedrive-team/filplus-info/settings"
	"github.com/filedrive-team/filplus-info/types"
	"github.com/filedrive-team/filplus-info/utils"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	logger "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

type Notary struct {
	Model
	NotaryName   string `json:"notary_name" gorm:"size:64"`           //notary name
	Organization string `json:"organization" gorm:"size:128"`         //organization
	Address      string `json:"address" gorm:"size:255;unique_index"` //notary address
	Location     string `json:"location" gorm:"size:64"`              //region
	Website      string `json:"website" gorm:"size:255"`              //website、social media
	Remark       string `json:"remark" gorm:"size:255"`               //remark

	GithubUser    string          `json:"github_user" gorm:"size:128"` //github account
	LastLoginTime *types.UnixTime `json:"last_login_time,omitempty"`   //the latest login time
}

func (Notary) TableName() string {
	return "notary"
}

func init() {
	autoMigrateModels = append(autoMigrateModels, &Notary{})
}

func GetNotaryList(offset, size int) (total int, list []*Notary, err error) {
	list = make([]*Notary, 0)
	stmt := db.Model(Notary{})
	var tmpTotal int64
	err = stmt.Count(&tmpTotal).Error
	if err != nil {
		err = errors.Wrap(err, "count notary table failed.")
		return
	}
	total = int(tmpTotal)
	err = stmt.Order("id desc").Offset(offset).Limit(size).Find(&list).Error
	if err != nil {
		err = errors.Wrap(err, "query notary table failed.")
		return
	}
	return
}

func NotaryList(params *types.PaginationParams) (*types.CommonList, error) {
	type NotaryItem struct {
		Notary
		TxnId          int64           `json:"txn_id"`
		Allowance      decimal.Decimal `json:"allowance"`
		GrantAllowance decimal.Decimal `json:"grant_allowance"`
	}

	offset, size := utils.PaginationHelper(params.Page, params.PageSize, settings.DefaultPageSize)

	stmtCount := `
select 
count(*)
from notary n
-- quotas held by notary
left join (
select
na.address,
na.allowance,
na.txn_id
from (
select na.address, max(txn_id) as txn_id 
from notary_allowance na group by na.address
) t2 left join notary_allowance na
on t2.txn_id=na.txn_id
) na 
on n.address = na.address 
-- notary has issued quotas
left join (
select 
ca.notary,
sum(ca.allowance) as grant_allowance 
from client_allowance ca 
group by ca.notary 
) t 
on n.address = t.notary
where n.deleted_at is null
`
	stmt := `
select 
n.*,
COALESCE(na.allowance,0) as allowance,
COALESCE(na.txn_id,0) as txn_id,
COALESCE(t.grant_allowance,0) as grant_allowance
from notary n
-- quotas held by notary
left join (
select
na.address,
na.allowance,
na.txn_id
from (
select na.address, max(txn_id) as txn_id 
from notary_allowance na group by na.address
) t2 left join notary_allowance na
on t2.txn_id=na.txn_id
) na 
on n.address = na.address 
-- notary has issued quotas
left join (
select 
ca.notary,
sum(ca.allowance) as grant_allowance 
from client_allowance ca 
group by ca.notary 
) t 
on n.address = t.notary
where n.deleted_at is null
order by grant_allowance desc
limit ? offset ?
`
	total := 0
	err := db.Raw(stmtCount).Scan(&total).Error
	if err != nil {
		return nil, err
	}
	var data []*NotaryItem
	err = db.Raw(stmt, size, offset).Scan(&data).Error
	if err != nil {
		return nil, err
	}
	res := &types.CommonList{
		Total: total,
		List:  data,
	}
	return res, nil
}

func InsertNotary(n *Notary) error {
	return db.Create(n).Error
}

func DeleteNotaryById(id uint) error {
	a := new(Notary)
	a.ID = id
	return db.Delete(a).Error
}

func TruncateNotary() error {
	return db.Exec("truncate notary;").Error
}

func InitNotaryList() error {
	var total int64
	err := db.Model(Notary{}).Count(&total).Error
	if err != nil {
		err = errors.Wrap(err, "count notary table failed.")
		return err
	}
	if total > 0 {
		logger.Info("notary table is not empty, skip it")
		return nil
	}
	list := []*Notary{
		{
			NotaryName:   "MathWallet",
			Location:     "Asia-GCN",
			Organization: "MathWallet",
			Address:      "f1snwd4w4y2fjwadkftcitcsnoerprlkrphn6kily",
			Website:      "mathwallet.org",
		},
		{
			NotaryName:   "Masaaki Nawatani",
			Location:     "Asia-GCN",
			Organization: "Blotocol Japan",
			Address:      "f1fh53sdaie3yi25qxwcqxpt5h4naex5ibdaffibi",
			Website:      "https://blotocol.com/",
		},
		{
			NotaryName:   "Keyko",
			Location:     "Europe",
			Organization: "Keyko",
			Address:      "f2m6qyszn47pd35bqs46on3ebwoo7fd242gog2gbq",
			Website:      "https://www.keyko.io",
		},
		{
			NotaryName:   "Julien NOEL",
			Location:     "Europe",
			Organization: "Julien NOEL",
			Address:      "f1wxhnytjmklj2czezaqcfl7eb4nkgmaxysnegwii",
			Website:      "s0nik42 on Filecoin Slack",
		},
		{
			NotaryName:   "Philipp Banhardt",
			Location:     "Europe",
			Organization: "Filecoin Foundation",
			Address:      "f1pns2ivst3kwrxatogpoucfk32ugebnn3medd73a",
			Website:      "https://fil.org/",
		},
		{
			NotaryName:   "Philipp Banhardt",
			Location:     "Europe",
			Organization: "Filecoin Foundation",
			Address:      "f1sdzgaqmitbvgktkklpuaxohg6nuhce5eyvwxhaa", // datacap已重置，该记录需软删除
			Website:      "@philipp on the Filecoin Slack",
		},
		{
			NotaryName:   "Steven Li",
			Location:     "GCN",
			Organization: "IPFS Force",
			Address:      "f1qoxqy3npwcvoqy7gpstm65lejcy7pkd3hqqekna",
			Website:      "@Steven on Filecoin Slack, @Steven004_Li on Twitter",
		},
		{
			NotaryName:   "Simon686",
			Location:     "GCN",
			Organization: "1475",
			Address:      "f1lwpw2bcv66pla3lpkcuzquw37pbx7ur4m6zvq2a",
			Website:      "1475ipfs.com, @simon686-1475 on Filecoin Slack",
		},
		{
			NotaryName:   "Fenbushi Capital",
			Location:     "GCN",
			Organization: "Fenbushi Capital",
			Address:      "f1yqydpmqb5en262jpottko2kd65msajax7fi4rmq",
			Website:      "https://fenbushi.vc",
		},
		{
			NotaryName:   "Neo Ge",
			Location:     "GCN",
			Organization: "IPFSMain",
			Address:      "f13k5zr6ovc2gjmg3lvd43ladbydhovpylcvbflpa",
			Website:      "@Neo Ge on Filecoin Slack",
		},
		{
			NotaryName:   "Jonathan Schwartz",
			Location:     "NA",
			Organization: "Infinite Scroll",
			Address:      "f3qqlzlsjxgy67wdwe5ade5ygk7omp6cnze3nr3aoxwtptjg3ar4i3w26p4rplnm7ppeeyjlwtxqawx2boioma",
			Website:      "infinitescroll.org",
			Remark:       "Glif Verifier",
		},
		{
			NotaryName:   "Jason Brozena",
			Location:     "NA",
			Organization: "Performive",
			Address:      "f13vzzb65gr7pjmb2vsfq666epq6lhdbanc4vfopq",
			Website:      "https://performive.com/",
		},
		{
			NotaryName:   "Andrew Hill",
			Location:     "NA",
			Organization: "Textile",
			Address:      "f2kb4izxsxu2jyyslzwmv2sfbrgpld56efedgru5i",
			Website:      "textile.io",
		},
		{
			NotaryName:   "Fleek",
			Location:     "NA",
			Organization: "Fleek",
			Address:      "f2lda563hy3q2oxnoo265elynct3u2ft27ae7ptfy",
			Website:      "fleek.co",
		},
		{
			NotaryName:   "Ann Shin",
			Location:     "NA",
			Organization: "XnMatrix",
			Address:      "f1yuz2twsllparyfqwslfiuxrc5wj4mfiflvnsw6a",
			Website:      "@ Dr.Ann Shin on Filecoin Slack",
		},
	}
	for _, item := range list {
		// record deletion processing
		if item.Address == "f1sdzgaqmitbvgktkklpuaxohg6nuhce5eyvwxhaa" {
			item.DeletedAt = &gorm.DeletedAt{
				Valid: true,
				Time:  time.Now(),
			}
		}
		err := InsertNotary(item)
		if err != nil {
			return err
		}
	}
	return nil
}

// Calculate the proportion based on whether it is allocated or not
func GetProportionOfAllowance() ([]*types.ProportionOfSomething, error) {
	sql := `
select 'Allocated' as name, (select sum(ca.allowance) from client_allowance ca) as value
union all
select 'Unallocated' as name, (select sum(na.allowance) from notary_allowance na)-(select sum(ca.allowance) from client_allowance ca) as value
`
	var data []*types.ProportionOfSomething
	err := db.Raw(sql).Scan(&data).Error
	return data, err
}

// Calculate the proportion of notary public quota based on region
func GetProportionOfAllowanceByLocation() ([]*types.ProportionOfSomething, error) {
	sql := `
select 
n.location as name,
COALESCE(sum(na.allowance),0) as value
from notary n
-- quota held by notary
left join (
select
na.address,
na.allowance,
na.txn_id
from (
select na.address, max(txn_id) as txn_id 
from notary_allowance na group by na.address
) t2 left join notary_allowance na
on t2.txn_id=na.txn_id
) na 
on n.address = na.address 
group by 1
`
	var data []*types.ProportionOfSomething
	err := db.Raw(sql).Scan(&data).Error
	return data, err
}
