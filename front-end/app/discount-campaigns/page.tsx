"use client"

import { useState, useEffect, useCallback } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Plus, Pencil, Trash2, Tag } from "lucide-react"
import { discountCampaignsApi, type DiscountCampaign } from "@/lib/api"
import { formatDate } from "@/lib/utils"

export default function DiscountCampaignsPage() {
  const [campaigns, setCampaigns] = useState<DiscountCampaign[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [isFormOpen, setIsFormOpen] = useState(false)
  const [editingCampaign, setEditingCampaign] = useState<DiscountCampaign | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<DiscountCampaign | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const [formName, setFormName] = useState("")
  const [formDiscount, setFormDiscount] = useState("")
  const [formStartDate, setFormStartDate] = useState("")
  const [formEndDate, setFormEndDate] = useState("")

  const fetchCampaigns = useCallback(async () => {
    setLoading(true)
    const result = await discountCampaignsApi.getAll()
    if (result.error) {
      setError(result.error)
    } else {
      setCampaigns(result.data)
    }
    setLoading(false)
  }, [])

  useEffect(() => {
    fetchCampaigns()
  }, [fetchCampaigns])

  const openCreate = () => {
    setEditingCampaign(null)
    setFormName("")
    setFormDiscount("")
    setFormStartDate("")
    setFormEndDate("")
    setIsFormOpen(true)
  }

  const openEdit = (campaign: DiscountCampaign) => {
    setEditingCampaign(campaign)
    setFormName(campaign.name)
    setFormDiscount(String(campaign.discount_percentage))
    setFormStartDate(campaign.start_date.slice(0, 10))
    setFormEndDate(campaign.end_date.slice(0, 10))
    setIsFormOpen(true)
  }

  const handleSubmit = async () => {
    if (!formName || !formDiscount || !formStartDate || !formEndDate) return
    setSubmitting(true)
    const payload = {
      name: formName,
      discount_percentage: parseFloat(formDiscount),
      start_date: formStartDate,
      end_date: formEndDate,
    }
    const result = editingCampaign
      ? await discountCampaignsApi.update(editingCampaign.id, payload)
      : await discountCampaignsApi.create(payload)
    if (result.error) {
      setError(result.error)
    } else {
      setIsFormOpen(false)
      fetchCampaigns()
    }
    setSubmitting(false)
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    const result = await discountCampaignsApi.delete(deleteTarget.id)
    if (result.error) {
      setError(result.error)
    } else {
      setDeleteTarget(null)
      fetchCampaigns()
    }
  }

  const isActive = (campaign: DiscountCampaign) => {
    const now = new Date()
    return new Date(campaign.start_date) <= now && now <= new Date(campaign.end_date)
  }

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-2">
          <Tag className="w-6 h-6 text-blue-600" />
          <h1 className="text-2xl font-bold">Discount Campaigns</h1>
        </div>
        <Button onClick={openCreate}>
          <Plus className="w-4 h-4 mr-2" />
          New Campaign
        </Button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-4">
          {error}
        </div>
      )}

      {loading ? (
        <div className="text-center py-12 text-gray-500">Loading campaigns...</div>
      ) : campaigns.length === 0 ? (
        <div className="text-center py-12 text-gray-500">
          No discount campaigns yet. Create one to auto-apply discounts during checkout.
        </div>
      ) : (
        <div className="space-y-4">
          {campaigns.map((campaign) => (
            <Card key={campaign.id}>
              <CardHeader className="pb-2">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-3">
                    <CardTitle className="text-lg">{campaign.name}</CardTitle>
                    <Badge variant={isActive(campaign) ? "default" : "secondary"}>
                      {isActive(campaign) ? "Active" : "Inactive"}
                    </Badge>
                  </div>
                  <div className="flex gap-2">
                    <Button variant="ghost" size="sm" onClick={() => openEdit(campaign)}>
                      <Pencil className="w-4 h-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="text-red-600 hover:text-red-700"
                      onClick={() => setDeleteTarget(campaign)}
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-2 gap-4 text-sm text-gray-600">
                  <div>
                    <span className="font-medium">Discount:</span>{" "}
                    <span className="text-green-600 font-bold">
                      {campaign.discount_percentage}%
                    </span>
                  </div>
                  <div>
                    <span className="font-medium">Period:</span>{" "}
                    {formatDate(new Date(campaign.start_date))} –{" "}
                    {formatDate(new Date(campaign.end_date))}
                  </div>
                  <div className="col-span-2">
                    <span className="font-medium">Products:</span>{" "}
                    {campaign.products && campaign.products.length > 0 ? (
                      <span className="flex flex-wrap gap-1 mt-1">
                        {campaign.products.map((p) => (
                          <Badge key={p.product_id} variant="outline" className="text-xs">
                            {p.name}
                          </Badge>
                        ))}
                      </span>
                    ) : (
                      <span className="text-gray-400">No products assigned</span>
                    )}
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Dialog open={isFormOpen} onOpenChange={setIsFormOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>
              {editingCampaign ? "Edit Campaign" : "New Discount Campaign"}
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="campaign-name">Campaign Name *</Label>
              <Input
                id="campaign-name"
                placeholder="e.g. Ramadan Sale"
                value={formName}
                onChange={(e) => setFormName(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="campaign-discount">Discount Percentage * (1–100)</Label>
              <Input
                id="campaign-discount"
                type="number"
                min="1"
                max="100"
                placeholder="e.g. 20"
                value={formDiscount}
                onChange={(e) => setFormDiscount(e.target.value)}
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="campaign-start">Start Date *</Label>
                <Input
                  id="campaign-start"
                  type="date"
                  value={formStartDate}
                  onChange={(e) => setFormStartDate(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="campaign-end">End Date *</Label>
                <Input
                  id="campaign-end"
                  type="date"
                  value={formEndDate}
                  onChange={(e) => setFormEndDate(e.target.value)}
                />
              </div>
            </div>
            {!editingCampaign && (
              <p className="text-xs text-gray-500">
                After creating, assign products to this campaign via the Products page.
              </p>
            )}
          </div>
          <DialogFooter className="mt-4">
            <Button variant="outline" onClick={() => setIsFormOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleSubmit}
              disabled={submitting || !formName || !formDiscount || !formStartDate || !formEndDate}
            >
              {submitting ? "Saving..." : editingCampaign ? "Update" : "Create"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={!!deleteTarget} onOpenChange={() => setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Campaign</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete &quot;{deleteTarget?.name}&quot;? This cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete} className="bg-red-600 hover:bg-red-700">
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
