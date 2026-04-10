"use client"

import { useState, useEffect } from "react"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { formatCurrency, formatDate } from "@/lib/utils"
import { Printer } from "lucide-react"
import ProductImage from "@/components/ProductImage"
import type { Transaction } from "@/lib/api"
import { tenantApi, type Tenant } from "@/lib/api" // Import tenantApi and Tenant type
import LoadingSpinner from "./LoadingSpinner" // Assuming you have this component
import ErrorMessage from "./ErrorMessage" // Assuming you have this component

interface TransactionDetailModalProps {
  isOpen: boolean
  onClose: () => void
  transaction: Transaction | null
}

export default function TransactionDetailModal({ isOpen, onClose, transaction }: TransactionDetailModalProps) {
  const [tenantDetails, setTenantDetails] = useState<Tenant | null>(null)
  const [loadingTenant, setLoadingTenant] = useState(true)
  const [tenantError, setTenantError] = useState<string | null>(null)

  useEffect(() => {
    if (isOpen) {
      const fetchTenant = async () => {
        setLoadingTenant(true)
        setTenantError(null)
        try {
          const result = await tenantApi.getTenantDetails()
          if (result.error) {
            setTenantError(result.error)
          } else {
            setTenantDetails(result.data)
          }
        } catch (err: any) {
          setTenantError(err.message || "Failed to load tenant details.")
        } finally {
          setLoadingTenant(false)
        }
      }
      fetchTenant()
    }
  }, [isOpen])

  const handlePrint = () => {
    // Small delay to ensure modal is fully rendered before printing
    setTimeout(() => {
      window.print()
    }, 100)
  }

  if (!transaction) return null

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-sm mx-4">
        <DialogHeader>
          <DialogTitle>Transaction Details</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          {/* Receipt Preview - Optimized for thermal printing */}
          <div id="transaction-receipt" className="receipt-container bg-white p-4 border rounded-lg text-sm">
            {/* Store Header */}
            <div className="text-center">
              {loadingTenant ? (
                <LoadingSpinner size="sm" className="my-4" />
              ) : tenantError ? (
                <ErrorMessage message={tenantError} className="my-4" />
              ) : tenantDetails ? (
                <>
                  <h2 className="font-bold text-lg">{tenantDetails.name}</h2>
                  <p className="text-xs text-gray-600">{tenantDetails.about}</p>
                  <p className="text-xs text-gray-600">{tenantDetails.address}</p>
                  <p className="text-xs text-gray-600">Telp: {tenantDetails.phone_number}</p>
                </>
              ) : (
                <>
                  <h2 className="font-bold text-lg">SAGITA STORE & SERVICE</h2>
                  <p className="text-xs text-gray-600">Jl. Srijaya Negara No.75</p>
                  <p className="text-xs text-gray-600">RW.2, Bukit Lama</p>
                  <p className="text-xs text-gray-600">Kec. Ilir Bar. I</p>
                  <p className="text-xs text-gray-600">Kota Palembang 30139</p>
                  <p className="text-xs text-gray-600">Telp: (0711) 123-4567</p>
                </>
              )}
            </div>

            {/* Transaction Info */}
            <div className="border-t border-b border-dashed py-2 my-2">
              <div className="flex justify-between text-xs">
                <span>TXN ID:</span>
                <span>{transaction.id}</span>
              </div>
              <div className="flex justify-between text-xs">
                <span>Date:</span>
                <span>{formatDate(new Date(transaction.created_at))}</span>
              </div>
              {transaction.user && (
                <div className="flex justify-between text-xs">
                  <span>Customer:</span>
                  <span>{transaction.user}</span>
                </div>
              )}
              {transaction.customer_email && (
                <div className="flex justify-between text-xs">
                  <span>Email:</span>
                  <span>{transaction.customer_email}</span>
                </div>
              )}
              {transaction.customer_phone && (
                <div className="flex justify-between text-xs">
                  <span>Phone:</span>
                  <span>{transaction.customer_phone}</span>
                </div>
              )}
            </div>

            {/* Items */}
            <div className="space-y-1 mb-2">
              {transaction.items.map((item, index) => (
                <div key={index}>
                  <div className="flex justify-between items-center">
                    <div className="flex items-center gap-1">
                      <ProductImage
                        src={item.product.image}
                        alt={item.product.name}
                        width={32}
                        height={32}
                        className="w-8 h-8"
                      />
                      <span className="text-xs font-medium">{item.product.name}</span>
                    </div>
                  </div>
                  <div className="flex justify-between text-xs">
                    <span>
                      {item.quantity} x {formatCurrency(item.price)}
                    </span>
                    <span className="text-right">{formatCurrency(item.price * item.quantity)}</span>
                  </div>
                  {item.discount_percentage && item.discount_percentage > 0 && (
                    <div className="text-xs text-green-600">
                      Discount: {item.discount_percentage}% off (campaign)
                    </div>
                  )}
                  {item.warranty_days && item.warranty_days > 0 && (
                    <div className="text-xs text-blue-600">
                      Warranty: {item.warranty_days} days
                    </div>
                  )}
                </div>
              ))}
            </div>

            {/* Total */}
            <div className="border-t border-dashed pt-2">
              <div className="flex justify-between font-bold">
                <span>TOTAL:</span>
                <span>{formatCurrency(transaction.total_price)}</span>
              </div>
              <div className="flex justify-between text-xs mt-1">
                <span>Payment:</span>
                <span className="uppercase font-bold">{transaction.payment_method}</span>
              </div>
            </div>

            {/* Notes */}
            {transaction.notes && (
              <div className="border-t border-dashed pt-2 mt-2">
                <div className="text-xs">
                  <span className="font-bold">Notes:</span>
                  <p className="mt-1">{transaction.notes}</p>
                </div>
              </div>
            )}

            {/* Footer */}
            <div className="text-center mt-4">
              <p className="text-xs">================================</p>
              <p className="text-xs font-bold">TERIMA KASIH</p>
              <p className="text-xs">THANK YOU FOR YOUR VISIT</p>
              <p className="text-xs">================================</p>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex space-x-3">
            <Button variant="outline" onClick={onClose} className="flex-1">
              Close
            </Button>
            <Button onClick={handlePrint} className="flex-1">
              <Printer className="w-4 h-4 mr-2" />
              Print Receipt
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
