"use client"

import { useState } from "react"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Card, CardContent } from "@/components/ui/card"
import { formatCurrency } from "@/lib/utils"
import ProductImage from "@/components/ProductImage"

interface CartItem {
  id: string
  name: string
  harga_jual: number // Updated field name
  quantity: number
  image?: string
}

interface CheckoutModalProps {
  isOpen: boolean
  onClose: () => void
  onConfirm: (
    customerName: string,
    paymentMethod: string,
    notes: string,
    customerEmail: string,
    customerPhone: string
  ) => void
  cart: CartItem[]
  total: number
}

export default function CheckoutModal({ isOpen, onClose, onConfirm, cart, total }: CheckoutModalProps) {
  const [customerName, setCustomerName] = useState("")
  const [paymentMethod, setPaymentMethod] = useState("")
  const [notes, setNotes] = useState("")
  const [customerEmail, setCustomerEmail] = useState("")
  const [customerPhone, setCustomerPhone] = useState("")

  const handleConfirm = () => {
    if (!paymentMethod) return
    onConfirm(customerName, paymentMethod, notes, customerEmail, customerPhone)
    setCustomerName("")
    setPaymentMethod("")
    setNotes("")
    setCustomerEmail("")
    setCustomerPhone("")
  }

  const handleClose = () => {
    setCustomerName("")
    setPaymentMethod("")
    setNotes("")
    setCustomerEmail("")
    setCustomerPhone("")
    onClose()
  }

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="max-w-md mx-4 max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Confirm Transaction</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          {/* Order Summary */}
          <Card>
            <CardContent className="p-4">
              <h3 className="font-semibold mb-3">Order Summary</h3>
              <div className="space-y-2 max-h-32 overflow-y-auto">
                {cart.map((item) => (
                  <div key={item.id} className="flex items-center justify-between text-sm">
                    <div className="flex items-center gap-2">
                      <ProductImage src={item.image} alt={item.name} width={32} height={32} className="w-8 h-8" />
                      <span>
                        {item.name} x{item.quantity}
                      </span>
                    </div>
                    <span>{formatCurrency(item.harga_jual * item.quantity)}</span>
                  </div>
                ))}
              </div>
              <div className="border-t pt-2 mt-3">
                <div className="flex justify-between font-bold">
                  <span>Total:</span>
                  <span className="text-green-600">{formatCurrency(total)}</span>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Customer Information */}
          <div className="space-y-2">
            <Label htmlFor="customer-name">User/Customer Name (Optional)</Label>
            <Input
              id="customer-name"
              type="text"
              placeholder="Enter user or customer name"
              value={customerName}
              onChange={(e) => setCustomerName(e.target.value)}
            />
          </div>

          {/* Customer Email */}
          <div className="space-y-2">
            <Label htmlFor="customer-email">Customer Email (Optional)</Label>
            <Input
              id="customer-email"
              type="email"
              placeholder="customer@example.com"
              value={customerEmail}
              onChange={(e) => setCustomerEmail(e.target.value)}
            />
          </div>

          {/* Customer Phone */}
          <div className="space-y-2">
            <Label htmlFor="customer-phone">Customer Phone (Optional)</Label>
            <Input
              id="customer-phone"
              type="tel"
              placeholder="e.g. 08123456789"
              value={customerPhone}
              onChange={(e) => setCustomerPhone(e.target.value)}
            />
          </div>

          {/* Payment Method */}
          <div className="space-y-2">
            <Label htmlFor="payment-method">Payment Method *</Label>
            <Select value={paymentMethod} onValueChange={setPaymentMethod}>
              <SelectTrigger>
                <SelectValue placeholder="Select payment method" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="cash">Cash</SelectItem>
                <SelectItem value="qris">QRIS</SelectItem>
                <SelectItem value="debit">Debit Card</SelectItem>
                <SelectItem value="credit">Credit Card</SelectItem>
                <SelectItem value="transfer">Bank Transfer</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Notes */}
          <div className="space-y-2">
            <Label htmlFor="notes">Notes (Optional)</Label>
            <Textarea
              id="notes"
              placeholder="Add any special instructions or notes..."
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              rows={3}
            />
          </div>

          {/* Action Buttons */}
          <div className="flex space-x-3 pt-4">
            <Button variant="outline" onClick={handleClose} className="flex-1">
              Cancel
            </Button>
            <Button onClick={handleConfirm} disabled={!paymentMethod} className="flex-1">
              Confirm Transaction
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
