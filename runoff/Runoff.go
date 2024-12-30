package Runoff

import (
	"demo2/Data"
	"math"
)

type Runoff struct {
	// ========模型参数======== //
	WM  float64 // 流域平均张力水容量（mm），不敏感，范围：120~200
	B   float64 // 张力水蓄水容量曲线方次，不敏感，范围：0.1~0.4
	IM  float64 // 不透水面积占全流域面积的比例，不敏感
	WUM float64 // 上层张力水容量（mm），敏感，范围：10~50
	WLM float64 // 下层张力水容量（mm），敏感，范围：60~90
	WDM float64 // 深层张力水容量（mm），等于 WM - WUM - WLM，不属于参数

	// ========模型状态======== //
	R   float64 // 总径流量（mm）
	RIM float64 // 不透水面积上的产流量（mm）
	W   float64 // 流域平均初始土壤含水量（mm）
	WU  float64 // 上层张力水蓄量（mm）
	WL  float64 // 下层张力水蓄量（mm）
	WD  float64 // 深层张力水蓄量（mm）
	WMM float64 // 包气带蓄水容量最大值（mm）
	A   float64 // 初始土壤含水量最大值（mm）
	EU  float64 // 上层蒸散发量（mm）
	EL  float64 // 下层蒸散发量（mm）
	ED  float64 // 深层蒸散发量（mm）
	EP  float64 // 流域蒸发能力（mm）
	PE  float64 // 净雨量（mm）

	// ========外部输入======== //
	P float64 // 降雨量（mm）
}

func (r *Runoff) SetParmameter(parameter *Data.Parameter) {
	r.WM = parameter.WM
	r.B = parameter.B
	r.IM = parameter.IM
	r.WUM = parameter.UM
	r.WLM = parameter.LM
}

func (r *Runoff) SetState(state *Data.State) {
	r.WU = state.WU
	r.WL = state.WL
	r.WD = state.WD
	r.W = state.W
	r.P = state.P
	r.EU = state.EU
	r.EL = state.EL
	r.ED = state.ED
	r.EP = state.EP
}

func (r *Runoff) UpdateState(state *Data.State) {
	state.WU = r.WU
	state.WL = r.WL
	state.WD = r.WD
	state.W = r.W
	state.R = r.R
	state.PE = r.PE
	state.RIM = r.RIM
}

func (r *Runoff) Calculate() {
	// ========计算产流========//
	r.WMM = (1 + r.B) / (1 - r.IM) * r.WM               // 包气带蓄水容量最大值，mm
	r.A = r.WMM * (1 - math.Pow(1-r.W/r.WM, 1/(1+r.B))) // 初始土壤含水量最大值，mm

	r.PE = r.P - r.EP
	if r.PE <= 1e-5 { // 这里认为净雨量小于1e-5时即为小于等于0
		r.R = 0.0
		r.RIM = 0.0 // 计算不透水面积上的产流量
	} else {
		if r.A+r.PE <= r.WMM {
			r.R = r.PE + r.W - r.WM + r.WM*math.Pow(1-(r.A+r.PE)/r.WMM, r.B+1)
		} else {
			r.R = r.PE - (r.WM - r.W)
		}
		r.RIM = r.PE * r.IM // 计算不透水面积上的产流量
	}

	// ========计算下一时段初土壤含水量========//
	r.WU = r.WU + r.P - r.EU - r.R
	r.WL = r.WL - r.EL
	r.WD = r.WD - r.ED
	if r.WD < 0 {
		r.WD = 0 // 防止深层张力水蓄量小于0
	}

	// 放置张力水蓄量超上限
	if r.WU > r.WUM {
		r.WL = r.WL + r.WU - r.WUM
		r.WU = r.WUM
	}
	if r.WL > r.WLM {
		r.WD = r.WD + r.WL - r.WLM
		r.WL = r.WLM
	}

	r.WDM = r.WM - r.WUM - r.WLM // 计算深层张力水容量
	if r.WD > r.WDM {
		r.WD = r.WDM
	}

	// 计算土壤含水量
	r.W = r.WU + r.WL + r.WD
}

func NewRunoff(wm, b, im, wum, wlm, wdm, r, rim, w, wu, wl, wd, wmm, a, eu, el, ed, ep, pe, p float64) *Runoff {
	return &Runoff{
		WM:  wm,
		B:   b,
		IM:  im,
		WUM: wum,
		WLM: wlm,
		WDM: wdm,
		R:   r,
		RIM: rim,
		W:   w,
		WU:  wu,
		WL:  wl,
		WD:  wd,
		WMM: wmm,
		A:   a,
		EU:  eu,
		EL:  el,
		ED:  ed,
		EP:  ep,
		PE:  pe,
		P:   p,
	}
}

func (r *Runoff) Destroy() {
	// 析构函数
}
