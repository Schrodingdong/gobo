package algorithm

import (
  "time"
  "log"
)

// Token Bucket algorithm
type Bucket struct {
  Size int                    // Num of tokens
  RefillDelay time.Duration   // Timestamped
  LastRefill time.Time
  BucketFill int              // Available tokens
}

func (b *Bucket) ConsumeToken() int {
  if b.BucketFill > 0 {
    b.BucketFill -= 1
  }
  return b.BucketFill
}

func (b *Bucket) refill() {
  b.BucketFill = b.Size
}

func RefillRoutine(bl *map[string]*Bucket) {
  for {
    t := time.Now()
    for k, b := range *bl {
      if b.RefillDelay < t.Sub(b.LastRefill) && b.Size > b.BucketFill{
        log.Printf("Refilling bucket for: %q\n", k)
        b.refill()
        b.LastRefill = t
      }
    } 
    time.Sleep(1 * time.Second)
  }
}
