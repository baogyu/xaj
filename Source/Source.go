package Source

import (
	"demo2/Data"
	"math"
)

type Source struct {
	// ========模型参数======== //
	SM float64 // 表层自由水蓄水容量（mm），敏感
	EX float64 // 表层自由水蓄水容量方次，不敏感，范围：1.0~1.5
	KG float64 // 表层自由水蓄水库对地下水的日出流系数，敏感
	KI float64 // 表层自由水蓄水库对壤中流的日出流系数，敏感

	// ========模型状态======== //
	R  float64 // 总径流量（mm）
	RS float64 // 地面径流（mm）
	RI float64 // 壤中流（mm）
	RG float64 // 地下径流（mm）
	PE float64 // 净雨量（mm）
	FR float64 // 本时段产流面积比例
	S0 float64 // 本时段初的自由水蓄量（mm）
	S  float64 // 本时段的自由水蓄量（mm）

	// ========辅助参数======== //
	M    float64 // 一天划分的计算时段数
	KID  float64 // 表层自由水蓄水库对壤中流的计算时段出流系数，敏感
	KGD  float64 // 表层自由水蓄水库对地下水的计算时段出流系数，敏感
	SMM  float64 // 全流域单点最大的自由水蓄水容量（mm）
	SMMF float64 // 产流面积上最大一点的自由水蓄水容量（mm）
	SMF  float64 // 产流面积上的平均自由水蓄水容量深（mm）
	AU   float64 // 对应平均蓄水深的最大蓄水深（mm）
	RSD  float64 // 计算步长地面径流（mm）
	RID  float64 // 计算步长壤中流（mm）
	RGD  float64 // 计算步长地下径流（mm）
	FR0  float64 // 上一时段产流面积比例

	// ========计算步长======== //
	N    int     // 计算时段分段数，每一段为计算步长
	Q    float64 // 每个计算步长内的净雨量（mm）
	KIDD float64 // 表层自由水蓄水库对壤中流的计算步长出流系数，敏感
	KGDD float64 // 表层自由水蓄水库对地下水的计算步长出流系数，敏感
	dt   float64 // 模型计算时段长（h）
}

func (s *Source) SetParmameter(parameter *Data.Parameter) {
	s.SM = parameter.SM
	s.EX = parameter.EX
	s.KG = parameter.KG
	s.KI = parameter.KI
}

func (s *Source) SetState(state *Data.State) {
	s.R = state.R
	s.PE = state.PE
	s.FR = state.FR
	s.S0 = state.S0
	s.dt = state.Dt
}

func (s *Source) UpdateState(state *Data.State) {
	state.RS = s.RS
	state.RI = s.RI
	state.RG = s.RG
	state.S0 = s.S
	state.FR = s.FR
}

func (s *Source) Calculate() {
	// 出流系数换算
	s.M = 24.0 / s.dt                                                // 一天划分的计算时段数
	s.KID = (1 - math.Pow(1-(s.KI+s.KG), 1.0/s.M)) / (1 + s.KG/s.KI) // 表层自由水蓄水库对壤中流的计算时段出流系数，敏感
	s.KGD = s.KID * s.KG / s.KI                                      // 表层自由水蓄水库对地下水的计算时段出流系数，敏感

	// 三分水源[4]
	if s.PE <= 1e-5 { // 净雨量小于等于0时,这里认为净雨量小于1e-5时即为小于等于0
		s.RS = 0
		s.RI = s.KID * s.S0 * s.FR // 当净雨量小于等于0时，消耗自由水蓄水库中的水
		s.RG = s.KGD * s.S0 * s.FR
		s.S = s.S0 * (1 - s.KID - s.KGD) // 更新下一时段初的自由水蓄量
	} else { // 净雨量大于0时
		s.SMM = s.SM * (1 + s.EX) // 全流域单点最大的自由水蓄水容量，mm
		s.FR0 = s.FR              // 上一时段产流面积比例
		s.FR = s.R / s.PE         // 计算本时段产流面积比例
		if s.FR > 1 {             // 如果FR由于小数误差而计算出大于1的情况，则强制置为1
			s.FR = 1
		}
		s.S = s.S0 * s.FR0 / s.FR

		s.N = int(s.PE/5.0) + 1   // N 为计算时段分段数，每一段为计算步长
		s.Q = s.PE / float64(s.N) // Q 是每个计算步长内的净雨量，mm

		s.KIDD = (1 - math.Pow(1-(s.KID+s.KGD), 1.0/float64(s.N))) / (1 + s.KGD/s.KID) // 表层自由水蓄水库对壤中流的计算步长出流系数，敏感
		s.KGDD = s.KIDD * s.KGD / s.KID                                                // 表层自由水蓄水库对地下水的计算步长出流系数，敏感

		s.RS = 0.0
		s.RI = 0.0
		s.RG = 0.0

		if s.EX == 0.0 {
			s.SMMF = s.SMM // EX等于0时，流域自由水蓄水容量分布均匀
		} else {
			s.SMMF = (1 - math.Pow(1-s.FR, 1.0/s.EX)) * s.SMM // 假定SMMF与产流面积FR及全流域上最大点的自由水蓄水容量SMM仍为抛物线分布
		}
		s.SMF = s.SMMF / (1.0 + s.EX)

		for i := 1; i <= s.N; i++ {
			if s.S > s.SMF {
				s.S = s.SMF
			}
			s.AU = s.SMMF * (1 - math.Pow(1-s.S/s.SMF, 1.0/(1+s.EX)))
			if s.Q+s.AU <= 0 {
				s.RSD = 0
				s.RID = 0
				s.RGD = 0
				s.S = 0
			} else {
				if s.Q+s.AU >= s.SMMF {
					s.RSD = (s.Q + s.S - s.SMF) * s.FR
					s.RID = s.SMF * s.KIDD * s.FR
					s.RGD = s.SMF * s.KGDD * s.FR
					s.S = s.SMF * (1 - s.KIDD - s.KGDD)
				} else {
					s.RSD = (s.S + s.Q - s.SMF + s.SMF*math.Pow(1-(s.Q+s.AU)/s.SMMF, 1+s.EX)) * s.FR
					s.RID = (s.S*s.FR + s.Q*s.FR - s.RSD) * s.KIDD
					s.RGD = (s.S*s.FR + s.Q*s.FR - s.RSD) * s.KGDD
					s.S = s.S + s.Q - (s.RSD+s.RID+s.RGD)/s.FR
				}
			}
			s.RS = s.RS + s.RSD
			s.RI = s.RI + s.RID
			s.RG = s.RG + s.RGD
		}
	}
}

func NewSource(sm, ex, kg, ki, r, rs, ri, rg, pe, fr, s0, s, m, kid, kgd, smm, smmf, smf, au, rsd, rid, rgd, fr0 float64, n int, q, kidd, kgdd, dt float64) *Source {
	return &Source{
		SM:   sm,
		EX:   ex,
		KG:   kg,
		KI:   ki,
		R:    r,
		RS:   rs,
		RI:   ri,
		RG:   rg,
		PE:   pe,
		FR:   fr,
		S0:   s0,
		S:    s,
		M:    m,
		KID:  kid,
		KGD:  kgd,
		SMM:  smm,
		SMMF: smmf,
		SMF:  smf,
		AU:   au,
		RSD:  rsd,
		RID:  rid,
		RGD:  rgd,
		FR0:  fr0,
		N:    n,
		Q:    q,
		KIDD: kidd,
		KGDD: kgdd,
		dt:   dt,
	}
}

func (s *Source) Destroy() {
	// 析构函数
}
