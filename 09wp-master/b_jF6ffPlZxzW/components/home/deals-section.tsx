"use client"

import { useState, useEffect } from "react"
import Image from "next/image"
import { Clock, ArrowRight } from "lucide-react"
import { Button } from "@/components/ui/button"

export interface Deal {
  id: number
  title: string
  image: string
  originalPrice: string
  salePrice: string
  discount: number
  endTime: string
}

function CountdownTimer({ endTime }: { endTime: string }) {
  const [timeLeft, setTimeLeft] = useState({
    days: 0,
    hours: 0,
    minutes: 0,
    seconds: 0,
  })

  useEffect(() => {
    const timer = setInterval(() => {
      const now = new Date().getTime()
      const end = new Date(endTime).getTime()
      const distance = end - now

      if (distance > 0) {
        setTimeLeft({
          days: Math.floor(distance / (1000 * 60 * 60 * 24)),
          hours: Math.floor((distance % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60)),
          minutes: Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60)),
          seconds: Math.floor((distance % (1000 * 60)) / 1000),
        })
      }
    }, 1000)

    return () => clearInterval(timer)
  }, [endTime])

  return (
    <div className="flex items-center gap-1 text-xs">
      <Clock className="h-3 w-3 text-primary" />
      <span className="text-muted-foreground">
        {timeLeft.days}天 {timeLeft.hours}:{String(timeLeft.minutes).padStart(2, "0")}:{String(timeLeft.seconds).padStart(2, "0")}
      </span>
    </div>
  )
}

export function DealsSection({ deals }: { deals: Deal[] }) {
  return (
    <section>
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-4">
          <h2 className="text-2xl font-bold text-foreground">限时特惠</h2>
          <span className="rounded-full bg-primary/20 px-3 py-1 text-xs font-medium text-primary">
            限时折扣
          </span>
        </div>
        <Button variant="ghost" className="gap-2 text-primary">
          查看全部特惠
          <ArrowRight className="h-4 w-4" />
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        {deals.map((deal) => (
          <div
            key={deal.id}
            className="group relative overflow-hidden rounded-xl bg-card border border-border transition-all duration-300 hover:border-primary/50"
          >
            <div className="flex gap-4 p-4">
              {/* Image */}
              <div className="relative h-24 w-24 flex-shrink-0 overflow-hidden rounded-lg">
                <Image
                  src={deal.image}
                  alt={deal.title}
                  fill
                  className="object-cover transition-transform duration-300 group-hover:scale-110"
                />
              </div>

              {/* Content */}
              <div className="flex flex-col justify-between flex-1 min-w-0">
                <div>
                  <h3 className="font-semibold text-foreground truncate group-hover:text-primary transition-colors">
                    {deal.title}
                  </h3>
                  <CountdownTimer endTime={deal.endTime} />
                </div>

                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="rounded bg-primary px-2 py-0.5 text-xs font-bold text-primary-foreground">
                      -{deal.discount}%
                    </span>
                    <div className="flex items-baseline gap-1">
                      <span className="text-xs text-muted-foreground line-through">
                        ¥{deal.originalPrice}
                      </span>
                      <span className="text-lg font-bold text-foreground">
                        ¥{deal.salePrice}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </section>
  )
}
