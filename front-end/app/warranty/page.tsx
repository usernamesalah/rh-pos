"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Shield, Search } from "lucide-react"
import { warrantyApi, type WarrantyResponse } from "@/lib/api"

function WarrantyCard({ warranty }: { warranty: WarrantyResponse }) {
  return (
    <Card className="mt-4">
      <CardHeader>
        <CardTitle className="text-base">
          Transaction: {warranty.transaction_id}
        </CardTitle>
        <p className="text-sm text-gray-500">
          Date: {new Date(warranty.transaction_date).toLocaleDateString("id-ID")}
        </p>
        {warranty.customer_name && (
          <p className="text-sm">Customer: {warranty.customer_name}</p>
        )}
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {warranty.items.map((item, i) => (
            <div key={i} className="border rounded-lg p-3">
              <div className="flex justify-between items-start">
                <div>
                  <p className="font-medium text-sm">{item.product_name}</p>
                  <p className="text-xs text-gray-500">Qty: {item.quantity}</p>
                </div>
                <Badge variant={item.is_active ? "default" : "secondary"}>
                  {item.is_active ? `${item.days_remaining} days left` : "Expired"}
                </Badge>
              </div>
              {item.warranty_days > 0 ? (
                <div className="mt-2 text-xs text-gray-600 space-y-1">
                  <p>Warranty: {item.warranty_days} days</p>
                  <p>
                    Valid:{" "}
                    {new Date(item.warranty_start).toLocaleDateString("id-ID")} –{" "}
                    {new Date(item.warranty_end).toLocaleDateString("id-ID")}
                  </p>
                </div>
              ) : (
                <p className="mt-2 text-xs text-gray-400">No warranty</p>
              )}
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

export default function WarrantyPage() {
  const [txId, setTxId] = useState("")
  const [phone, setPhone] = useState("")
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [singleResult, setSingleResult] = useState<WarrantyResponse | null>(null)
  const [listResult, setListResult] = useState<WarrantyResponse[] | null>(null)

  const searchById = async () => {
    if (!txId.trim()) return
    setLoading(true)
    setError(null)
    setSingleResult(null)
    const result = await warrantyApi.checkByTransactionId(txId.trim())
    if (result.error) {
      setError(result.error)
    } else {
      setSingleResult(result.data)
    }
    setLoading(false)
  }

  const searchByPhone = async () => {
    if (!phone.trim()) return
    setLoading(true)
    setError(null)
    setListResult(null)
    const result = await warrantyApi.searchByPhone(phone.trim())
    if (result.error) {
      setError(result.error)
    } else {
      setListResult(result.data ?? [])
    }
    setLoading(false)
  }

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col items-center justify-start pt-12 px-4">
      <div className="w-full max-w-lg">
        <div className="flex items-center gap-2 mb-6 justify-center">
          <Shield className="w-8 h-8 text-blue-600" />
          <h1 className="text-3xl font-bold text-gray-900">Warranty Check</h1>
        </div>
        <p className="text-center text-gray-500 mb-8 text-sm">
          Check your product warranty status by transaction ID or phone number.
        </p>

        <Tabs defaultValue="by-id">
          <TabsList className="w-full mb-4">
            <TabsTrigger value="by-id" className="flex-1">
              By Transaction ID
            </TabsTrigger>
            <TabsTrigger value="by-phone" className="flex-1">
              By Phone Number
            </TabsTrigger>
          </TabsList>

          <TabsContent value="by-id">
            <div className="space-y-3">
              <div className="space-y-2">
                <Label htmlFor="tx-id">Transaction ID</Label>
                <Input
                  id="tx-id"
                  placeholder="Enter transaction ID"
                  value={txId}
                  onChange={(e) => setTxId(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && searchById()}
                />
              </div>
              <Button onClick={searchById} disabled={loading || !txId.trim()} className="w-full">
                <Search className="w-4 h-4 mr-2" />
                {loading ? "Searching..." : "Check Warranty"}
              </Button>
            </div>
            {error && (
              <div className="mt-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded text-sm">
                {error}
              </div>
            )}
            {singleResult && <WarrantyCard warranty={singleResult} />}
          </TabsContent>

          <TabsContent value="by-phone">
            <div className="space-y-3">
              <div className="space-y-2">
                <Label htmlFor="phone">Phone Number</Label>
                <Input
                  id="phone"
                  type="tel"
                  placeholder="e.g. 08123456789"
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && searchByPhone()}
                />
              </div>
              <Button onClick={searchByPhone} disabled={loading || !phone.trim()} className="w-full">
                <Search className="w-4 h-4 mr-2" />
                {loading ? "Searching..." : "Search"}
              </Button>
            </div>
            {error && (
              <div className="mt-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded text-sm">
                {error}
              </div>
            )}
            {listResult !== null && listResult.length === 0 && (
              <p className="mt-4 text-center text-gray-500 text-sm">
                No warranties found for this phone number.
              </p>
            )}
            {listResult && listResult.map((w, i) => <WarrantyCard key={i} warranty={w} />)}
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}
