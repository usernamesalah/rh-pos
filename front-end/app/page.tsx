"use client"

import { useState, useEffect } from "react"
import { useRouter } from "next/navigation"
import Sidebar from "@/components/Sidebar"
import MobileNav from "@/components/MobileNav"
import CheckoutModal from "@/components/CheckoutModal"
import ReceiptModal from "@/components/ReceiptModal"
import ProductImage from "@/components/ProductImage"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "@/components/ui/sheet"
import { Search, Plus, Minus, Trash2, ShoppingCart, RefreshCw } from "lucide-react"
import { formatCurrency } from "@/lib/utils"
import { productsApi, transactionsApi, type Product, type CartItem, type Transaction } from "@/lib/api"
import LoadingSpinner from "@/components/LoadingSpinner"
import ErrorMessage from "@/components/ErrorMessage"

export default function POSPage() {
  const [searchTerm, setSearchTerm] = useState("")
  const [cart, setCart] = useState<CartItem[]>([])
  const [products, setProducts] = useState<Product[]>([])
  const [isCheckoutModalOpen, setIsCheckoutModalOpen] = useState(false)
  const [isReceiptModalOpen, setIsReceiptModalOpen] = useState(false)
  const [lastTransaction, setLastTransaction] = useState<Transaction | null>(null)
  const [isCartOpen, setIsCartOpen] = useState(false)
  const router = useRouter()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isAuthenticated, setIsAuthenticated] = useState(false)

  // Check authentication first
  useEffect(() => {
    const isLoggedIn = localStorage.getItem("isLoggedIn")
    const authToken = localStorage.getItem("authToken")

    if (!isLoggedIn || !authToken) {
      router.push("/login")
    } else {
      setIsAuthenticated(true)
    }
  }, [router])

  const fetchProducts = async () => {
    // Only fetch products if authenticated
    if (!isAuthenticated) {
      return
    }

    setLoading(true)
    setError(null)

    try {
      const result = await productsApi.getAll()
      if (result.error) {
        setError(result.error)
      } else {
        setProducts(result.data || [])
      }
    } catch (err: any) {
      setError(err.message || "Failed to load products. Please try again.")
    } finally {
      setLoading(false)
    }
  }

  // Only fetch products after authentication is confirmed
  useEffect(() => {
    if (isAuthenticated) {
      fetchProducts()
    }
  }, [isAuthenticated])

  const filteredProducts = products.filter(
    (product) =>
      product.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      product.sku.toLowerCase().includes(searchTerm.toLowerCase()),
  )

  const addToCart = (product: Product) => {
    setCart((prevCart) => {
      const existingItem = prevCart.find((item) => item.id === product.id)
      if (existingItem) {
        return prevCart.map((item) =>
          item.id === product.id ? { ...item, quantity: Math.min(item.quantity + 1, product.stock) } : item,
        )
      }
      return [...prevCart, { ...product, quantity: 1 }]
    })
  }

  const updateQuantity = (id: string, change: number) => {
    setCart((prevCart) =>
      prevCart
        .map((item) => {
          if (item.id === id) {
            const newQuantity = Math.max(0, Math.min(item.quantity + change, item.stock))
            return { ...item, quantity: newQuantity }
          }
          return item
        })
        .filter((item) => item.quantity > 0),
    )
  }

  const removeFromCart = (id: string) => {
    setCart((prevCart) => prevCart.filter((item) => item.id !== id))
  }

  const getTotalPrice = () => {
    return cart.reduce((total, item) => total + item.harga_jual * item.quantity, 0)
  }

  const handleCheckoutConfirm = async (
    customerName: string,
    paymentMethod: string,
    notes: string,
    customerEmail: string,
    customerPhone: string
  ) => {
    try {
      // Get the current user from localStorage
      const userStr = localStorage.getItem("user")
      const user = userStr ? JSON.parse(userStr) : { username: "admin" }

      const transactionData = {
        items: [...cart],
        total: getTotalPrice(),
        user: customerName || "", // Changed customerName to user
        paymentMethod,
        notes: notes || undefined,
        cashier: user.username,
        customer_name: customerName || undefined,
        customer_email: customerEmail || undefined,
        customer_phone: customerPhone || undefined,
      }

      const result = await transactionsApi.create(transactionData)

      if (result.error) {
        setError(result.error)
        return
      }

      setLastTransaction(result.data)
      setCart([])
      setIsCheckoutModalOpen(false)
      setIsReceiptModalOpen(true)
      setIsCartOpen(false)
    } catch (err: any) {
      setError(err.message || "Failed to create transaction. Please try again.")
    }
  }

  const CartContent = () => (
    <div className="h-full flex flex-col">
      <div className="flex-1 overflow-auto pb-4">
        {cart.length === 0 ? (
          <p className="text-gray-500 text-center py-8">Cart is empty</p>
        ) : (
          <div className="space-y-3">
            {cart.map((item) => (
              <Card key={item.id} className="p-3">
                <div className="flex items-center space-x-3">
                  <ProductImage src={item.image} alt={item.name} width={40} height={40} className="w-10 h-10" />
                  <div className="flex-1 min-w-0">
                    <h4 className="text-sm font-medium truncate">{item.name}</h4>
                    <p className="text-xs text-gray-500">{formatCurrency(item.harga_jual)}</p>
                  </div>
                  <div className="flex items-center space-x-1">
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => updateQuantity(item.id, -1)}
                      className="w-6 h-6 p-0"
                    >
                      <Minus className="w-3 h-3" />
                    </Button>
                    <span className="text-sm font-medium w-8 text-center">{item.quantity}</span>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => updateQuantity(item.id, 1)}
                      className="w-6 h-6 p-0"
                    >
                      <Plus className="w-3 h-3" />
                    </Button>
                    <Button
                      size="sm"
                      variant="destructive"
                      onClick={() => removeFromCart(item.id)}
                      className="w-6 h-6 p-0 ml-2"
                    >
                      <Trash2 className="w-3 h-3" />
                    </Button>
                  </div>
                </div>
                <div className="text-right mt-2">
                  <span className="text-sm font-semibold">{formatCurrency(item.harga_jual * item.quantity)}</span>
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>

      {cart.length > 0 && (
        <div className="sticky bottom-0 bg-white border-t pt-4 mt-auto">
          <div className="flex justify-between items-center mb-4">
            <span className="text-lg font-bold">Total:</span>
            <span className="text-xl font-bold text-green-600">{formatCurrency(getTotalPrice())}</span>
          </div>
          <Button onClick={() => setIsCheckoutModalOpen(true)} className="w-full" size="lg">
            Checkout
          </Button>
        </div>
      )}
    </div>
  )

  // Only render the main content if authenticated
  if (!isAuthenticated) {
    return (
      <div className="flex h-screen bg-gray-50">
        <div className="flex-1 flex items-center justify-center">
          <LoadingSpinner size="lg" />
          <p className="ml-2">Checking authentication...</p>
        </div>
      </div>
    )
  }

  if (loading) {
    return (
      <div className="flex h-screen bg-gray-50">
        <Sidebar />
        <div className="flex-1 flex items-center justify-center">
          <LoadingSpinner size="lg" />
        </div>
        <MobileNav />
      </div>
    )
  }

  return (
    <div className="flex h-screen bg-gray-50">
      <Sidebar />
      <div className="flex-1 flex flex-col lg:flex-row overflow-hidden">
        {/* Products Section */}
        <div className="flex-1 p-4 lg:p-6 overflow-auto pb-24 lg:pb-6">
          <div className="mb-6">
            <div className="flex justify-between items-center mb-4">
              <h1 className="text-xl lg:text-2xl font-bold text-gray-900">Point of Sale</h1>
              <Button variant="outline" onClick={fetchProducts} size="sm">
                <RefreshCw className="w-4 h-4 mr-2" />
                Refresh
              </Button>
            </div>
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
              <Input
                type="text"
                placeholder="Search products by name or SKU..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10"
              />
            </div>
          </div>

          {error && <ErrorMessage message={error} className="mb-4" />}

          {products.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-64">
              <p className="text-gray-500 mb-4">No products found</p>
              <Button onClick={fetchProducts}>
                <RefreshCw className="w-4 h-4 mr-2" />
                Refresh Products
              </Button>
            </div>
          ) : filteredProducts.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-64">
              <p className="text-gray-500 mb-4">No products match your search</p>
              <Button variant="outline" onClick={() => setSearchTerm("")}>
                Clear Search
              </Button>
            </div>
          ) : (
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-3 lg:gap-4 mb-20 lg:mb-0">
              {filteredProducts.map((product) => (
                <Card key={product.id} className="hover:shadow-md transition-shadow">
                  <CardContent className="p-3 lg:p-4">
                    <div className="relative pb-[75%] mb-3">
                      <ProductImage src={product.image} alt={product.name} className="absolute inset-0" fill />
                    </div>
                    <h3 className="font-semibold text-xs lg:text-sm mb-1 truncate">{product.name}</h3>
                    <p className="text-xs text-gray-500 mb-2">SKU: {product.sku}</p>
                    <div className="flex flex-wrap justify-between items-center mb-3 gap-2">
                      <span className="text-sm lg:text-base font-bold text-green-600">
                        {formatCurrency(product.harga_jual)}
                      </span>
                      <Badge
                        variant={product.stock > 10 ? "default" : "destructive"}
                        className="text-xs py-0 h-6 px-2 inline-flex items-center justify-center"
                      >
                        {product.stock} stock
                      </Badge>
                    </div>
                    <Button
                      onClick={() => addToCart(product)}
                      disabled={product.stock === 0}
                      className="w-full"
                      size="sm"
                    >
                      Add to Cart
                    </Button>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </div>

        {/* Desktop Cart Section */}
        <div className="hidden lg:block w-80 bg-white border-l border-gray-200 p-6 overflow-auto">
          <h2 className="text-xl font-bold mb-4">Cart</h2>
          <CartContent />
        </div>

        {/* Mobile Cart Sheet */}
        <Sheet open={isCartOpen} onOpenChange={setIsCartOpen}>
          <SheetHeader className="lg:hidden">
            {" "}
            {/* Added lg:hidden here */}
            <SheetTitle>Shopping Cart</SheetTitle>
          </SheetHeader>
          <SheetContent side="right" className="w-full sm:w-96 lg:hidden">
            <div className="mt-6 h-full">
              <CartContent />
            </div>
          </SheetContent>

          {/* Mobile Cart Button */}
          <div className="lg:hidden fixed bottom-20 right-4 z-40">
            <Button size="lg" className="rounded-full w-14 h-14 shadow-lg" onClick={() => setIsCartOpen(true)}>
              <ShoppingCart className="w-6 h-6" />
              {cart.length > 0 && (
                <Badge className="absolute -top-2 -right-2 w-6 h-6 rounded-full p-0 flex items-center justify-center">
                  {cart.reduce((sum, item) => sum + item.quantity, 0)}
                </Badge>
              )}
            </Button>
          </div>
        </Sheet>
      </div>

      <MobileNav />

      <CheckoutModal
        isOpen={isCheckoutModalOpen}
        onClose={() => setIsCheckoutModalOpen(false)}
        onConfirm={handleCheckoutConfirm}
        cart={cart}
        total={getTotalPrice()}
      />

      <ReceiptModal
        isOpen={isReceiptModalOpen}
        onClose={() => setIsReceiptModalOpen(false)}
        transaction={lastTransaction}
      />
    </div>
  )
}
