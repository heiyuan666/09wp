"use client"

import { useEffect, useState } from "react"
import { Search, User, ShoppingCart, Menu, X, Gamepad2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { SearchModal } from "@/components/home/search-modal"
import Link from "next/link"
import { fetchPublicNavMenus, type PublicNavMenuItem } from "@/lib/api/nav"
import { fetchPublicSystemConfig } from "@/lib/api/system"

export function Header() {
  const [isMenuOpen, setIsMenuOpen] = useState(false)
  const [isSearchOpen, setIsSearchOpen] = useState(false)
  const [nav, setNav] = useState<PublicNavMenuItem[]>([])
  const [token, setToken] = useState<string>("")
  const [siteTitle, setSiteTitle] = useState<string>("")
  const [logoUrl, setLogoUrl] = useState<string>("")

  const fallbackNav: Array<{ title: string; path: string }> = []

  // load nav menus from backend
  useEffect(() => {
    let mounted = true
    fetchPublicNavMenus("top_nav")
      .then((items) => {
        if (!mounted) return
        setNav(items)
      })
      .catch(() => {
        /* ignore */
      })
    return () => {
      mounted = false
    }
  }, [])

  // load site config (title/logo)
  useEffect(() => {
    let mounted = true
    fetchPublicSystemConfig()
      .then((cfg) => {
        if (!mounted) return
        setSiteTitle((cfg.site_title || "").trim())
        setLogoUrl((cfg.logo_url || "").trim())
      })
      .catch(() => {
        /* ignore */
      })
    return () => {
      mounted = false
    }
  }, [])

  // load auth token from localStorage (client-only)
  useEffect(() => {
    try {
      const t = localStorage.getItem("token") || ""
      if (t) setToken(t)
    } catch {
      /* ignore */
    }
  }, [])

  return (
    <header className="sticky top-0 z-40 border-b border-border bg-background/80 backdrop-blur-lg">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="flex h-16 items-center justify-between">
          {/* Logo */}
          <Link href="/" className="flex items-center gap-2">
            {logoUrl ? (
              <img src={logoUrl} alt={siteTitle || "logo"} className="h-8 w-8 rounded object-contain" />
            ) : (
              <Gamepad2 className="h-8 w-8 text-primary" />
            )}
            <span className="text-xl font-bold text-foreground">{siteTitle || "GameStore"}</span>
          </Link>

          {/* Desktop Nav */}
          <nav className="hidden md:flex items-center gap-8">
            {(nav.length ? nav : fallbackNav).map((item) => (
              <Link
                key={`${item.title}-${item.path}`}
                href={item.path || "/"}
                className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
              >
                {item.title}
              </Link>
            ))}
          </nav>

          {/* Actions */}
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="icon"
              className="text-muted-foreground hover:text-foreground"
              onClick={() => setIsSearchOpen(true)}
            >
              <Search className="h-5 w-5" />
            </Button>
            <Button variant="ghost" size="icon" className="text-muted-foreground hover:text-foreground relative">
              <ShoppingCart className="h-5 w-5" />
              <span className="absolute -top-1 -right-1 flex h-4 w-4 items-center justify-center rounded-full bg-primary text-[10px] font-bold text-primary-foreground">
                2
              </span>
            </Button>
            <Button variant="ghost" size="icon" className="text-muted-foreground hover:text-foreground" asChild>
              <Link href={token ? "/me" : "/login"}>
                <User className="h-5 w-5" />
              </Link>
            </Button>

            {/* Mobile menu button */}
            <Button 
              variant="ghost" 
              size="icon" 
              className="md:hidden text-muted-foreground hover:text-foreground"
              onClick={() => setIsMenuOpen(!isMenuOpen)}
            >
              {isMenuOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
            </Button>
          </div>
        </div>

        {/* Mobile Nav */}
        <div className={cn(
          "md:hidden overflow-hidden transition-all duration-300",
          isMenuOpen ? "max-h-48 pb-4" : "max-h-0"
        )}>
          <nav className="flex flex-col gap-2">
            {(nav.length ? nav : fallbackNav).map((item) => (
              <Link
                key={`${item.title}-${item.path}-m`}
                href={item.path || "/"}
                className="rounded-md px-3 py-2 text-sm font-medium text-muted-foreground hover:bg-secondary hover:text-foreground transition-colors"
                onClick={() => setIsMenuOpen(false)}
              >
                {item.title}
              </Link>
            ))}
          </nav>
        </div>
      </div>

      <SearchModal isOpen={isSearchOpen} onClose={() => setIsSearchOpen(false)} />
    </header>
  )
}
