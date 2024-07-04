package banta

import (
	"fmt"
	"log"
	"math"
	"time"
)

/*
缠论的主要构件：笔、线段、中枢、走势
笔和线段状态：CLInit/CLValid/CLDone
	最后一个笔可能会被移除，倒数第二个笔有效，倒数第三个笔确认完成
不同品种的不同周期对笔和线段适用不一。
	有些指数行情1分钟上可以用线段，但很多个股5分钟都不适用线段。
线段的完成由下一个线段的开始确认：
	一类：特征序列无缺口，力度小，新线段开始立刻完成当前线段
	二类：特征序列由缺口，力度大，新线段须由顶底分型确认有效，当前线段才完成

任务进度：
	* 笔：和原文略有不同，兼容一根极大K线包含几十根K线的情况
	* 线段：和原文略有不同，向后看最多7笔确定线段完成点
	* 中枢：只按最基本方法实现了
	* 趋势：未实现
*/

const (
	MinPenLen = 5
	CLInit    = 0 // 尚未确认有效
	CLValid   = 1 // 确认有效，尚未完成
	CLDone    = 2 // 确认有效，已完成，不会再改
)

const (
	EvtNew    = 1
	EvtChange = 2
	EvtRemove = 3
)

var (
	debugCL = true
)

type CPoint struct {
	Graph      *CGraph
	Dirt       float64 // 1顶分型，-1底分型
	BarId      int     // 此bar的序号
	Price      float64 // 价格
	StartPen   *CPen
	EndPen     *CPen
	StartSeg   *CSeg
	EndSeg     *CSeg
	StartTrend *CTrend
	EndTrend   *CTrend
	Next       *CPoint
}

type CTwoPoint struct {
	Graph *CGraph
	Start *CPoint
	End   *CPoint
	State int // 状态：0/1/2
}

type CPen struct {
	*CTwoPoint
	Dirt float64 // 1向上   -1向下
	Prev *CPen
	Next *CPen
}

/*
线段
*/
type CSeg struct {
	*CTwoPoint
	Dirt    float64      // 1向上  0不确定  -1向下
	Feas    [][2]float64 // 特征序列
	InForce bool         // 标记此线段是否必须以顶底分型确认
	Temp    bool         // 标记此线段是假设的，需要再次检查是否有效
	Prev    *CSeg
	Next    *CSeg
	Centre  *CCentre // 此线段所属中枢，只对中间部分线段设置；一个线段只可属于一个中枢
}

/*
中枢
*/
type CCentre struct {
	*CTwoPoint
	Overlap [2]float64 // 中枢重叠区间
	Range   [2]float64 // 中枢高低区间
	Dirt    float64
}

type CTrend struct {
	*CTwoPoint
	Dirt float64 // 1向上  0不确定  -1向下
	Prev *CTrend
	Next *CTrend
}

type CGraph struct {
	Pens     []*CPen
	Segs     []*CSeg
	Trends   []*CTrend
	Bars     []*Kline
	Centres  []*CCentre
	OnPoint  func(p *CPoint, evt int)
	OnPen    func(p *CPen, evt int)
	OnSeg    func(p *CSeg, evt int)
	OnCentre func(c *CCentre, evt int)
	BarNum   int     // 最后一个bar的序号，从1开始
	parseTo  int     // 解析到K线的位置，序号，非索引
	point    *CPoint // 最新一个点，用于查找最新的笔
}

type DrawLine struct {
	StartPos   int
	StartPrice float64
	StopPos    int
	StopPrice  float64
}

func setDebug(val bool) {
	debugCL = val
}

func (p *CPoint) Move(barId int, price float64) {
	if p.BarId != barId || p.Price != price {
		if debugCL {
			log.Printf("point move %s -> (%v, %.3f)\n", p.StrPoint(), barId, price)
		}
		c := p.Graph
		if p.StartPen != nil {
			pen := p.StartPen
			if barId >= pen.End.BarId {
				c.Remove(pen)
			}
		}
		if p.StartSeg != nil {
			seg := p.StartSeg
			if barId >= seg.End.BarId {
				c.Remove(seg)
			}
		}
		if p.StartTrend != nil {
			trend := p.StartTrend
			if barId >= trend.End.BarId {
				c.Remove(trend)
			}
		}
		p.BarId = barId
		p.Price = price
		if p.EndSeg != nil {
			p.EndSeg.CalcFeatures()
			if c.OnSeg != nil {
				c.OnSeg(p.EndSeg, EvtChange)
			}
		}
		if c.OnPoint != nil {
			c.OnPoint(p, EvtChange)
		}
		if c.OnPen != nil && p.EndPen != nil {
			c.OnPen(p.EndPen, EvtChange)
		}
	}
}

func (p *CPoint) PenTo(other *CPoint) *CPen {
	pen := p.Graph.NewPen(p, other)
	p.StartPen = pen
	other.EndPen = pen
	return pen
}

/*
State 分型状态，只返回CLInit/CLDone分别表示无效、有效
*/
func (p *CPoint) State() int {
	if p.Next != nil {
		return CLDone
	}
	if p.Graph.parseTo-p.BarId < MinPenLen-1 {
		return CLInit
	}
	lastB := p.Graph.Bar(p.Graph.parseTo)
	if p.Dirt > 0 && lastB.High < p.Price {
		// 顶分型，最新价格低于顶分型最高价，有效
		return CLDone
	} else if p.Dirt < 0 && lastB.Low > p.Price {
		// 底分型，最新价格高于底分型最低价，有效
		return CLDone
	}
	// 中继分型，无效
	return CLInit
}

/*
PensTo 从当前点出发，到指定点的所有笔
*/
func (p *CPoint) PensTo(end *CPoint) []*CPen {
	res := make([]*CPen, 0, 4)
	pt := p
	for pt.StartPen != nil {
		res = append(res, pt.StartPen)
		pt = pt.StartPen.End
		if pt == end {
			return res
		}
	}
	return nil
}

func (p *CPoint) Clear() {
	c := p.Graph
	if p.StartPen != nil {
		c.Remove(p.StartPen)
	}
	if p.EndPen != nil {
		c.Remove(p.EndPen)
	}
	if p.StartSeg != nil {
		c.Remove(p.StartSeg)
	}
	if p.EndSeg != nil {
		c.Remove(p.EndSeg)
	}
	if p.StartTrend != nil {
		c.Remove(p.StartTrend)
	}
	if p.EndTrend != nil {
		c.Remove(p.EndTrend)
	}
}

func (p *CPoint) StrPoint() string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("(%v, %.3f)", p.BarId, p.Price)
}

func (p *CTwoPoint) ToFeature() [2]float64 {
	var a, b = p.Start.Price, p.End.Price
	if a > b {
		a, b = b, a
	}
	return [2]float64{a, b}
}

func (p *CTwoPoint) String() string {
	return fmt.Sprintf("%s %s", p.Start.StrPoint(), p.End.StrPoint())
}

/*
Clear 删除笔对前后的关联
*/
func (p *CPen) Clear() {
	if p == nil {
		return
	}
	if p.Start != nil {
		p.Start.StartPen = nil
		p.Start = nil
	}
	if p.End != nil {
		p.End.EndPen = nil
		p.End = nil
	}
	if p.Prev != nil {
		p.Prev.Next = p.Next
		p.Prev = nil
	}
	if p.Next != nil {
		p.Next.Prev = p.Prev
		p.Next = nil
	}
}

/*
AddFeature 将笔作为特征序列
*/
func (s *CSeg) AddFeature(f [2]float64) {
	s.Feas = append(s.Feas, f)
	s.checkInForce()
}

func (s *CSeg) SetLastFea(f [2]float64) {
	s.Feas[len(s.Feas)-1] = f
	s.checkInForce()
}

func (s *CSeg) checkInForce() {
	if s.InForce && len(s.Feas) > 1 {
		fea2 := s.Feas[len(s.Feas)-2]
		fea := s.Feas[len(s.Feas)-1]
		if intersect(fea2, fea) == 0 {
			s.InForce = false
		}
	}
}

func (s *CSeg) SetEnd(p *CPoint) {
	if s.End != nil {
		s.End.EndSeg = nil
	}
	if debugCL {
		log.Printf("set seg %s - %s end: %s\n", s.Start.StrPoint(), s.End.StrPoint(), p.StrPoint())
	}
	s.End = p
	p.EndSeg = s
	s.State = CLValid
	if s.Dirt == 0 {
		if p.Price > s.Start.Price {
			s.Dirt = 1
		} else {
			s.Dirt = -1
		}
	}
	if s.Graph.OnSeg != nil {
		s.Graph.OnSeg(s, EvtChange)
	}
	if s.Prev == nil {
		return
	}
	// 如果笔数量至少7笔，则检查中间是否有突破线段起点的，如有设为起点
	pen := s.End.EndPen
	pens := make([]*CPen, 0, 9)
	pens = append(pens, pen)
	for pen.Start != s.Start && pen.Prev != nil {
		pen = pen.Prev
		pens = append(pens, pen)
	}
	if len(pens) >= 7 {
		entPrice := s.Start.Price
		var cmpVal = float64(0)
		var cmpPoint *CPoint
		for i := 2; i < len(pens)-1; i += 2 {
			pen = pens[i]
			curDiff := (pen.Start.Price - entPrice) * s.Dirt
			if curDiff < cmpVal {
				// 发现更高(低)的点，改为两线段的交界点
				cmpVal = curDiff
				cmpPoint = pen.Start
			}
		}
		if cmpPoint != nil {
			s.Start.StartSeg = nil
			s.Start = cmpPoint
			cmpPoint.StartSeg = s
			s.Prev.SetEnd(cmpPoint)
			s.Prev.CalcFeatures()
			s.CalcFeatures()
		}
	}
}

/*
CalcFeatures 重新计算线段的所有特征序列，当线段结束点修改时调用此方法
*/
func (s *CSeg) CalcFeatures() {
	oldNum := len(s.Feas)
	secondPen := s.Start.StartPen.Next
	s.Feas = [][2]float64{secondPen.ToFeature()}
	feaPen := secondPen.Next.Next // 下一个特征序列所在笔
	for {
		if feaPen == nil || feaPen.Next == nil || feaPen.Next.State < CLDone {
			break
		}
		lastFea := s.Feas[len(s.Feas)-1]
		fea2 := feaPen.ToFeature() // 下一个特征序列
		// 更新下一个nPen
		feaPen = feaPen.Next.Next
		startDiff := fea2[0] - lastFea[0]
		endDiff := fea2[1] - lastFea[1]
		startEndSame := startDiff * endDiff
		if startEndSame > 0 {
			// 新特征序列高低点都更高(低)
			if startDiff*s.Dirt > 0 {
				// 特征序列与当前线段方向一致，延续线段
				s.AddFeature(fea2)
				continue
			}
		}
		// 特征序列包含、或方向不一致。按线段方向更新最后一个特征序列
		if s.Temp {
			s.SetLastFea(fea2)
		} else {
			s.SetLastFea(mergeFea(s.Dirt, lastFea, fea2))
		}
	}
	if debugCL {
		log.Printf("seg %s, update feas, %v -> %v\n", s.String(), oldNum, len(s.Feas))
	}
}

func (s *CSeg) IsValid() bool {
	if !s.Temp {
		return true
	}
	if len(s.Feas) < 1 {
		return false
	}
	f := s.Feas[len(s.Feas)-1]
	if s.Dirt > 0 {
		return s.End.Price > f[1]
	} else {
		return s.End.Price < f[0]
	}
}

func (s *CSeg) fireDone() {
	if s.State < CLDone || s.Prev == nil || s.Prev.Prev == nil || len(s.Graph.Segs) < 4 {
		return
	}
	fea := s.ToFeature()
	pv2 := s.Prev.Prev
	if pv2.Centre != nil && s.Dirt*pv2.Centre.Dirt > 0 {
		cen := pv2.Centre
		lap1 := intersect(cen.Overlap, s.Prev.ToFeature())
		lap2 := intersect(cen.Overlap, fea)
		if max(lap1, lap2) > 0.2 && lap2 > 0.1 {
			// 最新两段与中枢重合超过20%，依旧在中枢范围内
			s.Centre = cen
			s.Prev.Centre = cen
			cen.SetEnd(s.End)
			return
		}
	}
	// 找到3个没有中枢的连续完成线段，检查是否重合
	seg := s
	segs := make([]*CSeg, 0, 3)
	for seg != nil && seg.Centre == nil {
		segs = append(segs, seg)
		seg = seg.Prev
		if len(segs) == 3 {
			break
		}
	}
	if len(segs) >= 3 {
		fea3 := segs[2].ToFeature()
		if intersect(fea, fea3) > 0 {
			// 有重叠部分，构成中枢
			overlap := [2]float64{max(fea[0], fea3[0]), min(fea[1], fea3[1])}
			if intersect(overlap, segs[1].ToFeature()) > 0.3 {
				// 中枢区间占第二段至少30%
				cen := &CCentre{
					CTwoPoint: &CTwoPoint{
						Graph: s.Graph,
						Start: segs[2].Start,
						End:   s.End,
					},
					Overlap: overlap,
					Range:   [2]float64{min(fea[0], fea3[0]), max(fea[1], fea3[1])},
					Dirt:    segs[2].Dirt,
				}
				s.Graph.Centres = append(s.Graph.Centres, cen)
				segs[2].Centre = cen
				segs[1].Centre = cen
				s.Centre = cen
			}
		}
	}
}

func (s *CSeg) Clear() {
	if s == nil {
		return
	}
	if s.Start != nil {
		s.Start.StartSeg = nil
		s.Start = nil
	}
	if s.End != nil {
		s.End.EndSeg = nil
		s.End = nil
	}
	if s.Prev != nil {
		s.Prev.Next = s.Next
		s.Prev = nil
	}
	if s.Next != nil {
		s.Next.Prev = s.Prev
		s.Next = nil
	}
}

func (c *CCentre) SetEnd(p *CPoint) {
	pt := c.End
	for pt.StartPen != nil {
		pt = pt.StartPen.End
		price := pt.Price
		if price > c.Range[1] {
			c.Range[1] = price
		} else if price < c.Range[0] {
			c.Range[0] = price
		}
		if pt == p {
			break
		}
	}
	c.End = p
}

func (t *CTrend) Clear() {
	if t == nil {
		return
	}
	if t.Start != nil {
		t.Start.StartTrend = nil
		t.Start = nil
	}
	if t.End != nil {
		t.End.EndTrend = nil
		t.End = nil
	}
	if t.Prev != nil {
		t.Prev.Next = t.Next
		t.Prev = nil
	}
	if t.Next != nil {
		t.Next.Prev = t.Prev
		t.Next = nil
	}
}

/*
NewPoint 返回新点，不添加到Points
*/
func (c *CGraph) NewPoint(dirt float64, price float64, barId int) *CPoint {
	if barId == 0 {
		barId = c.parseTo
	}
	return &CPoint{
		Graph: c,
		Dirt:  dirt,
		BarId: barId,
		Price: price,
	}
}

func (c *CGraph) NewPen(a, b *CPoint) *CPen {
	pen := &CPen{
		CTwoPoint: &CTwoPoint{
			Graph: c,
			Start: a,
			End:   b,
		},
	}
	if a.Price < b.Price {
		pen.Dirt = 1
	} else {
		pen.Dirt = -1
	}
	return pen
}

/*
AddPen 添加到Pens
*/
func (c *CGraph) AddPen(pen *CPen) {
	if debugCL {
		log.Printf("new %v pen %s\n", pen.Dirt, pen.String())
	}
	if len(c.Pens) > 0 {
		last := c.Pens[len(c.Pens)-1]
		if last.Dirt*pen.Dirt >= 0 {
			panic(fmt.Sprintf("up and down must appear alternately, %v -> %v", last.Dirt, pen.Dirt))
		}
		if last.State == CLInit {
			// 倒数第二个笔，确认有效，但结束位置可能变
			last.State = CLValid
			if c.OnPen != nil {
				c.OnPen(last, EvtChange)
			}
		}
		if last.Prev != nil && last.Prev.State < CLDone {
			// 倒数第三个笔，确认完成，不会变动
			last.Prev.State = CLDone
			if c.OnPen != nil {
				c.OnPen(last.Prev, EvtChange)
			}
		}
		last.Next = pen
		pen.Prev = last
	}
	c.Pens = append(c.Pens, pen)
	if c.OnPen != nil {
		c.OnPen(pen, EvtNew)
	}
	if c.OnPoint != nil {
		c.OnPoint(pen.End, EvtNew)
	}
	c.buildSegs()
}

func (c *CGraph) AddSeg(seg *CSeg) {
	if len(c.Segs) > 0 {
		last := c.Segs[len(c.Segs)-1]
		if last.Dirt*seg.Dirt >= 0 {
			panic(fmt.Sprintf("up and down must appear alternately, %v -> %v", last.Dirt, seg.Dirt))
		}
		last.Next = seg
		seg.Prev = last
	}
	c.Segs = append(c.Segs, seg)
	if c.OnSeg != nil {
		c.OnSeg(seg, EvtNew)
	}
}

/*
Bar 返回指定序号的Bar
*/
func (c *CGraph) Bar(v int) *Kline {
	lastIdx := len(c.Bars) - (c.BarNum - v) - 1
	if lastIdx >= 0 && lastIdx < len(c.Bars) {
		return c.Bars[lastIdx]
	}
	return nil
}

func (c *CGraph) Remove(o interface{}) bool {
	if pen, ok := o.(*CPen); ok {
		if debugCL {
			log.Printf("remove pen: %s\n", pen.String())
		}
		pen.Clear()
		c.Pens = RemoveFromArr(c.Pens, pen, 1)
		if c.OnPen != nil {
			c.OnPen(pen, EvtRemove)
		}
		if c.OnPoint != nil {
			c.OnPoint(pen.End, EvtRemove)
		}
	} else if seg, ok := o.(*CSeg); ok {
		if debugCL {
			log.Printf("remove seg: %s\n", seg.String())
		}
		seg.Clear()
		c.Segs = RemoveFromArr(c.Segs, seg, 1)
		if c.OnSeg != nil {
			c.OnSeg(seg, EvtRemove)
		}
	} else if trend, ok := o.(*CTrend); ok {
		if debugCL {
			log.Printf("remove trend: %s\n", trend.String())
		}
		trend.Clear()
		c.Trends = RemoveFromArr(c.Trends, trend, 1)
	} else {
		return false
	}
	return true
}

func (c *CGraph) AddBar(e *BarEnv) *CGraph {
	c.Bars = append(c.Bars, &Kline{
		Time:   e.TimeStart,
		Open:   e.Open.Get(0),
		High:   e.High.Get(0),
		Low:    e.Low.Get(0),
		Close:  e.Close.Get(0),
		Volume: e.Volume.Get(0),
		Info:   e.Info.Get(0),
	})
	c.BarNum = e.BarNum
	return c
}

func (c *CGraph) AddBars(barId int, bars ...*Kline) *CGraph {
	c.Bars = append(c.Bars, bars...)
	c.BarNum = barId + len(bars) - 1
	return c
}

func (c *CGraph) Parse() {
	barLen := len(c.Bars)
	if barLen < 3 {
		return
	}
	var pv, pv2 *CPoint
	if len(c.Pens) > 0 {
		pen := c.Pens[len(c.Pens)-1]
		pv, pv2 = pen.End, pen.Start
	} else {
		pv = c.point
	}
	var stPrice = c.Bars[0].Close
	var gapPrice = float64(0) // 上个点的压力或支撑价格，新点必须完全突破此线
	if pv != nil {
		stPrice = pv.Price
		b := c.Bar(pv.BarId)
		gapPrice = b.Low*0.5 + b.High*0.5
	}
	// 第一个待解析bar的索引
	startIdx := barLen - (c.BarNum - 1 - c.parseTo)
	if c.parseTo == 0 {
		//从第二个开始处理（因为需要前一个bar）
		startIdx = 1
		c.parseTo = 1
	}
	// 前一个节点的价格，用于判断当前大概趋势，节点更新时更新
	for i := startIdx; i < barLen-1; i++ {
		c.parseTo += 1
		pb, cb, nb := c.Bar(i), c.Bar(i+1), c.Bar(i+2)
		vhigh, vlow := cb.High, cb.Low
		// 这里是为了处理突然一个特别长的bar包含前面所有bar的情况，从前面价格的向上向下距离确定取高还是取低
		isUp := vhigh > stPrice && (vhigh-stPrice >= stPrice-vlow)
		var pt *CPoint
		if isUp && vhigh > max(pb.High, nb.High) {
			pt = c.NewPoint(1, vhigh, c.parseTo)
		} else if !isUp && vlow < min(pb.Low, nb.Low) {
			pt = c.NewPoint(-1, vlow, c.parseTo)
		} else {
			continue
		}
		if pv != nil {
			if pv.Dirt*pt.Dirt > 0 {
				//多个连续同方向的分型，只保留最后一个
				if (pt.Price-pv.Price)*pv.Dirt > 0 {
					pv.Move(pt.BarId, pt.Price)
					stPrice = pv.Price
					b := c.Bar(pv.BarId - 1)
					gapPrice = b.Low*0.5 + b.High*0.5
				}
				continue
			} else if pt.Dirt < 0 && cb.High > gapPrice || pt.Dirt > 0 && cb.Low < gapPrice {
				// 必须突破上一个点的gapPrice，才算有效的逆向点
				continue
			} else if pt.BarId-pv.BarId < MinPenLen-1 {
				// 不满足最少数量，不能构成笔
				if pv2 != nil && math.Abs(pt.Price-pv2.Price) > math.Abs(pv.Price-pv2.Price) {
					//如果当前点变化幅度超过前一个笔，则删除前一个笔，将倒数第二个笔结束设为当前，避免极值点丢失
					c.Remove(pv.EndPen)
					pv2.Move(pt.BarId, pt.Price)
					pv = pv2
					if pv.EndPen != nil {
						pv2 = pv.EndPen.Start
					} else {
						pv2 = nil
					}
					stPrice = pv.Price
					b := c.Bar(pv.BarId - 1)
					gapPrice = b.Low*0.5 + b.High*0.5
				}
				continue
			}
		}
		if debugCL {
			dateStr := time.Unix(cb.Time/1000, 0).Format("2006-01-02 15:04:05")
			log.Printf("[%s] point %s, up: %v\n", dateStr, pt.StrPoint(), isUp)
		}
		if pv != nil {
			pen := pv.PenTo(pt)
			c.AddPen(pen)
		}
		pv2 = pv
		pv = pt
		stPrice = pv.Price
		gapPrice = cb.Low*0.5 + cb.High*0.5
	}
	c.point = pv
}

/*
判断两个特征序列是否相交（特征序列已排序）
*/
func intersect(fa, fb [2]float64) float64 {
	a1, a2 := fa[0], fa[1]
	b1, b2 := fb[0], fb[1]
	if a1 <= b2 && b1 <= a2 {
		overlap := min(a2, b2) - max(a1, b1)
		return overlap / (b2 - b1)
	}
	return 0
}

/*
按给定方向合并特征序列
*/
func mergeFea(dirt float64, a, b [2]float64) [2]float64 {
	var res = a
	firstSub := b[0] - a[0]
	secondSub := b[1] - a[1]
	if dirt*firstSub > 0 {
		res[0] = b[0]
	}
	if dirt*secondSub > 0 {
		res[1] = b[1]
	}
	return res
}

/*
buildSegs 尝试从笔构建线段
*/
func (c *CGraph) buildSegs() {
	var seg *CSeg
	if len(c.Segs) > 0 {
		seg = c.Segs[len(c.Segs)-1]
	} else {
		seg = c.buildFirstSeg(0)
	}
	if seg == nil {
		return
	}
	for {
		feaPen := seg.End.StartPen
		if feaPen == nil || feaPen.Next == nil || feaPen.Next.State < CLDone {
			// 要求后面两笔都已完成，不会再变，才尝试更新线段，避免未完成的笔更新错误
			break
		}
		feaNum := len(seg.Feas)
		p2, p3 := feaPen.Next, feaPen.Next.Next
		// 检查新特征序列与前一个是否对齐（高低点都更高/低，且与线段方向一致）
		fea := seg.Feas[feaNum-1]
		fea2 := feaPen.ToFeature() // 下一个特征序列
		startDiff := fea2[0] - fea[0]
		feaAlign := startDiff*(fea2[1]-fea[1]) > 0                  // 特征序列高低点都更高(低)
		feaDirtOk := startDiff*seg.Dirt > 0                         // 特征序列方向与线段方向一致
		lastPenExtend := (p2.End.Price-seg.End.Price)*seg.Dirt >= 0 // 是否符合奇数笔更高(低)
		mayNewSeg, newSeg, newIsTemp, feaP3Align := false, false, false, false
		// 检查新的三个笔是否能构成新有效线段
		if p3 != nil {
			startSub := p3.Start.Price - feaPen.Start.Price
			endSub := p3.End.Price - feaPen.End.Price
			feaP3Align = startSub*endSub > 0
			if feaP3Align && startSub*feaPen.Dirt > 0 {
				// 后三笔能形成新线段
				newSeg = true
			}
			if newSeg && !lastPenExtend {
				// 不满足奇数笔高低点都更高(低)，且后三笔能形成新线段时，假设为新线段开始
				mayNewSeg = true
			}
		}
		mergeToPrev := func(end *CPoint) {
			// 将最后一个线段合并到前一个线段
			prev := seg.Prev
			prev.SetEnd(end)
			prev.CalcFeatures()
			c.Remove(seg)
			seg = prev
		}
		var newSegEnd *CPoint
		if !mayNewSeg {
			var checkCurFenXing = false // 是否应检查当前顶底分型
			if feaAlign {
				// 新特征序列高低点都更高(低)
				if !feaDirtOk && (seg.InForce && !newSeg || !seg.IsValid()) {
					// 要求顶底分型，下一个方向不一致，且未出现新有效线段，则回退
					mergeToPrev(seg.End.EndPen.Start)
					continue
				}
				if feaDirtOk && lastPenExtend {
					// 特征序列与当前线段方向一致，且符合奇数笔创新高(低)，延续线段
					seg.SetEnd(p2.End)
					seg.AddFeature(fea2)
				} else if p3 == nil {
					// 新线段的第三笔尚未出现，退出
					break
				} else if newSeg && !feaDirtOk {
					// 出现新线段，且特征序列方向与当前线段方向不一致
					mayNewSeg = true
				} else {
					// 方向一致，未创新高(低); 或未出现新线段
					checkCurFenXing = true
				}
			} else if lastPenExtend {
				// 特征序列包含，且最后一笔符合高低点都更高(低)，按线段方向更新最后一个特征序列
				seg.SetLastFea(mergeFea(seg.Dirt, fea, fea2))
				seg.SetEnd(p2.End)
			} else if p3 != nil {
				// 特征序列包含，但最后一笔不符合高低点都更高(低)
				if newSeg {
					// 且出现新有效线段，视为新线段
					mayNewSeg = true
				} else {
					// 没有新有效线段，检查从当前结束点是否能构成顶底分型，如果能构成，则形成新线段
					checkCurFenXing = true
				}
			} else {
				// 第三笔尚未出现
				break
			}
			if checkCurFenXing {
				// 检查当前是否符合顶底分型，如符合从当前创建新线段
				if p3.Next == nil || p3.Next.State < CLDone {
					break
				}
				p4 := p3.Next
				p3fea := mergeFea(seg.Dirt, fea2, p3.ToFeature())
				colIdx := 0
				if seg.Dirt > 0 {
					colIdx = 1
				}
				if (p4.End.Price-p3fea[colIdx])*(p3fea[colIdx]-fea[colIdx]) > 0 {
					// 从feaPen不构成顶底分型，故合并特征序列
					seg.AddFeature(p3fea)
					seg.SetEnd(p4.End)
				} else {
					// 顶底分型不需要严格满足，只需第三元素起点和第一元素起点相对第二元素一样即可
					mayNewSeg = true
					newIsTemp = true
					if p4.Next != nil {
						newSegEnd = p4.Next.End
					} else {
						break
					}
				}
			}
		}
		if mayNewSeg {
			// 新线段开始
			if newSegEnd == nil {
				newSegEnd = p3.End
			}
			hasFenXing := (newSeg || newIsTemp) && (feaNum > 1 || feaAlign)
			if seg.InForce && !hasFenXing || !seg.IsValid() {
				// seg必须顶底分型确认；
				// 不满足要求，移除此线段，合并到前一个线段
				mergeToPrev(newSegEnd)
				continue
			}
			var next = &CSeg{
				CTwoPoint: &CTwoPoint{
					Graph: c,
					Start: seg.End,
					End:   newSegEnd,
					State: CLValid,
				},
				Dirt: -seg.Dirt,
				Prev: seg,
				Temp: newIsTemp,
			}
			next.Start.StartSeg = next
			seg.Next = next
			// 当前线段最后特征序列是否有缺口，如有缺口，新线段必须以顶底分型结束
			var requireForce = intersect(fea, fea2) == 0
			if requireForce {
				next.State = CLInit
				next.InForce = true
				seg.State = CLValid
				if c.OnSeg != nil {
					c.OnSeg(seg, EvtChange)
				}
				if seg.Prev != nil && seg.Prev.State < CLDone {
					seg.Prev.State = CLDone
					seg.Prev.fireDone()
					if c.OnSeg != nil {
						c.OnSeg(seg.Prev, EvtChange)
					}
				}
			} else {
				seg.State = CLDone
				seg.fireDone()
			}
			c.AddSeg(next)
			next.CalcFeatures()
			if debugCL {
				log.Printf("new seg: %s\n", next.String())
			}
			seg = next
		}
	}
}

/*
tryBuildSeg 从指定点向后构建最小线段，仅用于构建第一个线段
*/
func (c *CGraph) buildFirstSeg(penIdx int) *CSeg {
	if len(c.Pens)-penIdx < 3 {
		return nil
	}
	pen := c.Pens[penIdx]
	seg := &CSeg{
		CTwoPoint: &CTwoPoint{
			Graph: c,
			Start: pen.Start,
			State: CLInit,
		},
		Dirt: 0,
	}
	seg.Start.StartSeg = seg
	initIdx := penIdx
	var pen3 *CPen
	for penIdx+2 < len(c.Pens) {
		pen3 = c.Pens[penIdx+2]
		startDiff := pen3.Start.Price - pen.Start.Price
		endDiff := pen3.End.Price - pen.End.Price
		startEndSame := startDiff * endDiff
		if startEndSame > 0 {
			// 第三笔相比第一笔，高低都更高(低)
			if startDiff*pen.Dirt > 0 {
				// 且与第一笔的方向相同，线段有效，记录特征序列
				seg.SetEnd(pen3.End)
				seg.AddFeature(pen3.Prev.ToFeature())
				c.AddSeg(seg)
				if debugCL {
					log.Printf("first %v seg %s\n", seg.Dirt, seg.String())
				}
				return seg
			} else {
				// 与第一笔方向不同，第一笔不能作为线段起始
				if debugCL {
					log.Printf("first pen can not build seg: %s\n", pen.String())
				}
				return c.buildFirstSeg(initIdx + 1)
			}
		}
		// 有包含关系，合并，然后继续向后看
		pen = c.NewPen(pen.Start, pen3.End)
		pen.Dirt = pen3.Dirt
		penIdx += 2
	}
	// 所有奇数笔都有包含关系，无法构成线段
	return nil
}

func (c *CGraph) Dump() []*DrawLine {
	res := make([]*DrawLine, 0)
	//for _, p := range c.Pens {
	//	start, inPrice := p.Start.BarId, p.Start.Price
	//	stop, outPrice := p.End.BarId, p.End.Price
	//	res = append(res, &DrawLine{start, inPrice, stop, outPrice})
	//}
	for _, p := range c.Segs {
		start, inPrice := p.Start.BarId, p.Start.Price
		stop, outPrice := p.End.BarId, p.End.Price
		res = append(res, &DrawLine{start, inPrice, stop, outPrice})
	}
	for _, cen := range c.Centres {
		start, stop := cen.Start.BarId, cen.End.BarId
		topPrice, btmPrice := cen.Overlap[1], cen.Overlap[0]
		res = append(res, &DrawLine{start, topPrice, stop, topPrice})
		res = append(res, &DrawLine{start, btmPrice, stop, btmPrice})
	}
	return res
}
