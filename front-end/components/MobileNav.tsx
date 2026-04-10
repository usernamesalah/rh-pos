"use client"

import { usePathname } from "next/navigation"
import Link from "next/link"
import { cn } from "@/lib/utils"
import { ShoppingCart, Package, BarChart3, Receipt, Tag } from "lucide-react"

const navigation = [
  { name: "POS", href: "/", icon: ShoppingCart },
  { name: "Products", href: "/products", icon: Package },
  { name: "Transactions", href: "/transactions", icon: Receipt },
  { name: "Reports", href: "/reports", icon: BarChart3 },
  { name: "Campaigns", href: "/discount-campaigns", icon: Tag },
]

export default function MobileNav() {
  const pathname = usePathname()

  return (
    <div className="lg:hidden fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 z-30">
      <div className="grid grid-cols-5">
        {navigation.map((item) => {
          const isActive = pathname === item.href
          return (
            <Link
              key={item.name}
              href={item.href}
              className={cn(
                "flex flex-col items-center justify-center py-3 px-2 text-xs font-medium transition-colors",
                isActive ? "text-blue-600 bg-blue-50" : "text-gray-600 hover:text-gray-900",
              )}
            >
              <item.icon className={cn("w-5 h-5 mb-1", isActive ? "text-blue-600" : "text-gray-400")} />
              {item.name}
            </Link>
          )
        })}
      </div>
    </div>
  )
}
