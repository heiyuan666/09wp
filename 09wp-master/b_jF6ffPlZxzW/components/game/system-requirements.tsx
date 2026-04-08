"use client"

import { useState } from "react"
import { cn } from "@/lib/utils"

interface SystemSpec {
  os: string
  processor: string
  memory: string
  graphics: string
  storage: string
  directX?: string
}

interface SystemRequirementsProps {
  windows: SystemSpec
  mac: SystemSpec
  linux: SystemSpec
}

export function SystemRequirements({ windows, mac, linux }: SystemRequirementsProps) {
  const [activeTab, setActiveTab] = useState<"windows" | "mac" | "linux">("windows")

  const specs = activeTab === "windows" ? windows : activeTab === "mac" ? mac : linux

  return (
    <div className="space-y-4">
      <h3 className="text-xl font-semibold text-foreground">系统需求</h3>
      
      {/* Tabs */}
      <div className="flex gap-2">
        <button
          onClick={() => setActiveTab("windows")}
          className={cn(
            "rounded-md px-4 py-2 text-sm font-medium transition-colors",
            activeTab === "windows"
              ? "bg-primary text-primary-foreground"
              : "bg-secondary text-secondary-foreground hover:bg-secondary/80"
          )}
        >
          Win
        </button>
        <button
          onClick={() => setActiveTab("mac")}
          className={cn(
            "rounded-md px-4 py-2 text-sm font-medium transition-colors",
            activeTab === "mac"
              ? "bg-primary text-primary-foreground"
              : "bg-secondary text-secondary-foreground hover:bg-secondary/80"
          )}
        >
          Mac
        </button>
        <button
          onClick={() => setActiveTab("linux")}
          className={cn(
            "rounded-md px-4 py-2 text-sm font-medium transition-colors",
            activeTab === "linux"
              ? "bg-primary text-primary-foreground"
              : "bg-secondary text-secondary-foreground hover:bg-secondary/80"
          )}
        >
          Linux
        </button>
      </div>

      {/* Specs */}
      <div className="rounded-lg bg-card p-4 space-y-3">
        <div className="grid grid-cols-[100px_1fr] gap-2 text-sm">
          <span className="text-muted-foreground">操作系统</span>
          <span className="text-foreground">{specs.os}</span>
        </div>
        <div className="grid grid-cols-[100px_1fr] gap-2 text-sm">
          <span className="text-muted-foreground">处理器</span>
          <span className="text-foreground">{specs.processor}</span>
        </div>
        <div className="grid grid-cols-[100px_1fr] gap-2 text-sm">
          <span className="text-muted-foreground">内存</span>
          <span className="text-foreground">{specs.memory}</span>
        </div>
        <div className="grid grid-cols-[100px_1fr] gap-2 text-sm">
          <span className="text-muted-foreground">显卡</span>
          <span className="text-foreground">{specs.graphics}</span>
        </div>
        <div className="grid grid-cols-[100px_1fr] gap-2 text-sm">
          <span className="text-muted-foreground">存储空间</span>
          <span className="text-foreground">{specs.storage}</span>
        </div>
        {specs.directX && (
          <div className="grid grid-cols-[100px_1fr] gap-2 text-sm">
            <span className="text-muted-foreground">DirectX</span>
            <span className="text-foreground">{specs.directX}</span>
          </div>
        )}
      </div>
    </div>
  )
}
