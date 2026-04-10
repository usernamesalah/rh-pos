"use client"

import { useRouter, usePathname } from "next/navigation"
import Link from "next/link"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { ShoppingCart, Package, BarChart3, LogOut, Receipt, KeyRound, Tag } from "lucide-react"

const navigation = [
  { name: "POS", href: "/", icon: ShoppingCart },
  { name: "Products", href: "/products", icon: Package },
  { name: "Transactions", href: "/transactions", icon: Receipt },
  { name: "Reports", href: "/reports", icon: BarChart3 },
  { name: "Campaigns", href: "/discount-campaigns", icon: Tag },
  { name: "Settings", href: "/settings", icon: KeyRound },
]

export default function Sidebar() {
  const router = useRouter()
  const pathname = usePathname()

  const handleLogout = () => {
    localStorage.removeItem("isLoggedIn")
    router.push("/login")
  }

  return (
    <div className="hidden lg:block w-64 bg-white border-r border-gray-200">
      <div className="flex flex-col h-full">
        {/* Logo */}
        <div className="flex items-center justify-center h-16 px-4 border-b border-gray-200">
          <div className="flex items-center space-x-2">
            <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
              <ShoppingCart className="w-5 h-5 text-white" />
            </div>
            <span className="text-xl font-bold text-gray-900">POS System</span>
          </div>
        </div>

        {/* Navigation */}
        <nav className="flex-1 px-4 py-6 space-y-2">
          {navigation.map((item) => {
            const isActive = pathname === item.href
            return (
              <Link
                key={item.name}
                href={item.href}
                className={cn(
                  "flex items-center px-4 py-3 text-sm font-medium rounded-lg transition-colors",
                  isActive
                    ? "bg-blue-50 text-blue-700 border-r-2 border-blue-700"
                    : "text-gray-600 hover:bg-gray-50 hover:text-gray-900",
                )}
              >
                <item.icon className="w-5 h-5 mr-3" />
                {item.name}
              </Link>
            )
          })}
        </nav>

        {/* Logout */}
        <div className="p-4 border-t border-gray-200">
          <Button variant="outline" className="w-full justify-start" onClick={handleLogout}>
            <LogOut className="w-4 h-4 mr-2" />
            Logout
          </Button>
        </div>
      </div>
    </div>
  )
}
