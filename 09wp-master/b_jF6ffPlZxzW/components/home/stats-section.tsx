"use client"

import { Gamepad2, Download, Users, HardDrive } from "lucide-react"

const stats = [
  {
    icon: Gamepad2,
    value: "10,000+",
    label: "游戏资源",
  },
  {
    icon: Download,
    value: "500万+",
    label: "总下载量",
  },
  {
    icon: Users,
    value: "100万+",
    label: "注册用户",
  },
  {
    icon: HardDrive,
    value: "50TB+",
    label: "存储空间",
  },
]

export function StatsSection() {
  return (
    <section className="rounded-2xl bg-gradient-to-br from-card via-card to-primary/5 border border-border p-8">
      <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
        {stats.map((stat, index) => {
          const Icon = stat.icon
          return (
            <div key={index} className="text-center">
              <div className="inline-flex items-center justify-center h-12 w-12 rounded-full bg-primary/10 mb-4">
                <Icon className="h-6 w-6 text-primary" />
              </div>
              <div className="text-2xl md:text-3xl font-bold text-foreground mb-1">
                {stat.value}
              </div>
              <div className="text-sm text-muted-foreground">
                {stat.label}
              </div>
            </div>
          )
        })}
      </div>
    </section>
  )
}
