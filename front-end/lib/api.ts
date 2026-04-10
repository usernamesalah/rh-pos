// API service layer with fallback to mock data
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001/api"

interface ApiResponse<T> {
  status: string
  message: string
  data: T
  pagination?: {
    total: number
    page: number
    limit: number
  }
}

// Update Product interface to match API response
interface Product {
  id: string
  name: string
  harga_jual: number // Changed from price
  harga_modal: number // Changed from costPrice
  stock: number
  image: string // This will now store the final image URL
  sku: string
  description?: string
  created_at?: string
  updated_at?: string
  // Removed upload_url as it's no longer used in the new flow
}

interface CartItem extends Product {
  quantity: number
}

// Updated TransactionItem to match API response
interface TransactionItem {
  price: number
  product: Product
  product_id: string
  quantity: number
  warranty_days?: number
  discount_percentage?: number
  campaign_id?: string
}

// Updated Transaction interface to match API response
interface Transaction {
  id: string
  items: TransactionItem[]
  total_price: number
  discount: number
  notes: string
  payment_method: string
  user: string
  customer_name?: string
  customer_email?: string
  customer_phone?: string
  created_at: string
  updated_at: string
}

// Interface for creating a new transaction
interface CreateTransactionData {
  items: Array<{
    product_id: string
    quantity: number
    price: number
    warranty_days?: number
  }>
  total_price: number
  discount?: number
  notes?: string
  payment_method: string
  user?: string
  customer_name?: string
  customer_email?: string
  customer_phone?: string
}

// Update the User interface to match the actual API response
interface User {
  id: string
  username: string
  role: string
  created_at?: string
  updated_at?: string
}

// Updated SalesData interface to match the actual API response
interface SalesDetail {
  created_at: string
  id: string
  product_id: number
  product_name: string
  total: number
  total_price: number
  updated_at: string
}

interface SalesReport {
  average_transaction: number
  details: SalesDetail[]
  items_sold: number
  total_revenue: number
}

// New Tenant interface
interface Tenant {
  name: string
  about: string
  address: string
  phone_number: string
}

interface DiscountCampaignProduct {
  id: string
  product_id: string
  name: string
  sku: string
}

interface DiscountCampaign {
  id: string
  name: string
  discount_percentage: number
  start_date: string
  end_date: string
  products: DiscountCampaignProduct[]
  created_at: string
  updated_at: string
}

interface CreateDiscountCampaignData {
  name: string
  discount_percentage: number
  start_date: string
  end_date: string
  product_ids?: string[]
}

interface UpdateDiscountCampaignData {
  name?: string
  discount_percentage?: number
  start_date?: string
  end_date?: string
}

interface WarrantyItem {
  product_name: string
  quantity: number
  warranty_days: number
  warranty_start: string
  warranty_end: string
  is_active: boolean
  days_remaining: number
}

interface WarrantyResponse {
  transaction_id: string
  transaction_date: string
  customer_name?: string
  customer_email?: string
  customer_phone?: string
  items: WarrantyItem[]
}

// Products API - Updated to use the actual API response
export const productsApi = {
  getAll: async () => {
    try {
      const token = localStorage.getItem("authToken")

      // Don't make the API call if there's no token
      if (!token) {
        console.warn("No auth token found, skipping product fetch")
        return {
          data: [],
          loading: false,
          error: "Authentication required. Please log in.",
        }
      }

      const headers: HeadersInit = {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      }

      const response = await fetch(`${API_BASE_URL}/api/products`, {
        headers,
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        const errorMessage = errorData.message || `HTTP error! status: ${response.status}`
        throw new Error(errorMessage)
      }

      const result = await response.json()

      if (result.status !== "success") {
        throw new Error(result.message || "Failed to fetch products")
      }

      return {
        data: result.data,
        loading: false,
        error: null,
        pagination: result.pagination,
      }
    } catch (error: any) {
      console.error("Error fetching products:", error.message)
      return {
        data: [],
        loading: false,
        error: error.message || "Failed to fetch products",
      }
    }
  },

  create: async (product: Omit<Product, "id" | "created_at" | "updated_at" | "image">, imageFile?: File) => {
    try {
      const token = localStorage.getItem("authToken")
      const headers: HeadersInit = {
        "Content-Type": "application/json",
      }

      if (token) {
        headers["Authorization"] = `Bearer ${token}`
      }

      // 1. Send product data without the image initially
      const productDataToSend = { ...product }

      const response = await fetch(`${API_BASE_URL}/api/products`, {
        method: "POST",
        headers,
        body: JSON.stringify(productDataToSend),
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        const errorMessage = errorData.message || `HTTP error! status: ${response.status}`
        throw new Error(errorMessage)
      }

      const result = await response.json()

      if (result.status !== "success") {
        throw new Error(result.message || "Failed to create product")
      }

      const createdProduct: Product = result.data
      let imageUploadError: string | null = null

      // 2. If an image file is provided, upload it to the new dedicated endpoint
      if (imageFile) {
        try {
          const formData = new FormData()
          formData.append("image", imageFile) // Append the file with the name 'image'

          console.log("Attempting to upload image to:", `${API_BASE_URL}/api/products/${createdProduct.id}/image`)
          console.log("Image file type:", imageFile.type)

          const uploadResponse = await fetch(`${API_BASE_URL}/api/products/${createdProduct.id}/image`, {
            method: "POST", // Use POST for the new image upload endpoint
            headers: {
              Authorization: `Bearer ${token}`, // Include authorization for this endpoint
              // 'Content-Type': 'multipart/form-data' is automatically set by fetch when using FormData
            },
            body: formData,
          })

          if (!uploadResponse.ok) {
            const uploadErrorText = await uploadResponse.text().catch(() => uploadResponse.statusText)
            imageUploadError = `Failed to upload image: ${uploadResponse.status} - ${uploadErrorText}`
            console.error("Image upload failed:", imageUploadError)
          } else {
            // Optionally, if the image upload API returns the updated product with the image URL,
            // you can update createdProduct here. For now, we assume the initial product creation
            // returns the product with a placeholder or default image, and the image upload
            // updates it on the backend. We'll refetch products to ensure consistency.
            console.log("Image uploaded successfully.")
          }
        } catch (uploadError: any) {
          imageUploadError = `Error during image upload: ${uploadError.message}`
          console.error("Error during image upload:", uploadError.message)
        }
      }

      // Return the created product data. The image URL will be updated on a subsequent fetch.
      return {
        data: createdProduct,
        loading: false,
        error: imageUploadError, // Return image upload error here
      }
    } catch (error: any) {
      console.error("Error creating product (initial API call or general):", error.message)
      return {
        data: null,
        loading: false,
        error: error.message || "Failed to create product",
      }
    }
  },

  update: async (id: string, product: Partial<Product>, imageFile?: File) => {
    try {
      const token = localStorage.getItem("authToken")
      const headers: HeadersInit = {
        "Content-Type": "application/json",
      }

      if (token) {
        headers["Authorization"] = `Bearer ${token}`
      }

      // 1. Send product data without the image initially (or with existing image URL if no new file)
      const productDataToSend: Partial<Product> = { ...product }
      if (imageFile) {
        delete productDataToSend.image // Remove image from payload if a new file is being uploaded
      }

      const response = await fetch(`${API_BASE_URL}/api/products/${id}`, {
        method: "PUT",
        headers,
        body: JSON.stringify(productDataToSend),
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        const errorMessage = errorData.message || `HTTP error! status: ${response.status}`
        throw new Error(errorMessage)
      }

      const result = await response.json()

      if (result.status !== "success") {
        throw new Error(result.message || "Failed to update product")
      }

      const updatedProduct: Product = result.data
      let imageUploadError: string | null = null

      // 2. If a new image file is provided, upload it to the new dedicated endpoint
      if (imageFile) {
        try {
          const formData = new FormData()
          formData.append("image", imageFile) // Append the file with the name 'image'

          console.log("Attempting to upload image to:", `${API_BASE_URL}/api/products/${updatedProduct.id}/image`)
          console.log("Image file type:", imageFile.type)

          const uploadResponse = await fetch(`${API_BASE_URL}/api/products/${updatedProduct.id}/image`, {
            method: "POST", // Use POST for the new image upload endpoint
            headers: {
              Authorization: `Bearer ${token}`, // Include authorization for this endpoint
            },
            body: formData,
          })

          if (!uploadResponse.ok) {
            const uploadErrorText = await uploadResponse.text().catch(() => uploadResponse.statusText)
            imageUploadError = `Failed to upload image: ${uploadResponse.status} - ${uploadErrorText}`
            console.error("Image upload failed:", imageUploadError)
          } else {
            console.log("Image uploaded successfully.")
          }
        } catch (uploadError: any) {
          imageUploadError = `Error during image upload: ${uploadError.message}`
          console.error("Error during image upload:", uploadError.message)
        }
      }

      // Return the updated product data. The image URL will be updated on a subsequent fetch.
      return {
        data: updatedProduct,
        loading: false,
        error: imageUploadError, // Return image upload error here
      }
    } catch (error: any) {
      console.error("Error updating product (initial API call or general):", error.message)
      return {
        data: null,
        loading: false,
        error: error.message || "Failed to update product",
      }
    }
  },

  addStock: async (id: string, stock: number) => {
    try {
      const token = localStorage.getItem("authToken")
      const headers: HeadersInit = {
        "Content-Type": "application/json",
      }

      if (token) {
        headers["Authorization"] = `Bearer ${token}`
      }

      const response = await fetch(`${API_BASE_URL}/api/products/${id}/stock`, {
        method: "PUT",
        headers,
        body: JSON.stringify({ stock }),
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const result = await response.json()

      if (result.status !== "success") {
        throw new Error(result.message || "Failed to add stock")
      }

      return {
        data: result.data,
        loading: false,
        error: null,
      }
    } catch (error: any) {
      console.error("Error adding stock:", error.message)
      // Fallback to mock data is removed as per previous instructions
      return {
        data: null,
        loading: false,
        error: error.message || "Failed to add stock",
      }
    }
  },
}

// Transactions API - Updated to use the actual API response
export const transactionsApi = {
  getAll: async (filters?: {
    startDate?: string
    endDate?: string
    productId?: string
    cashier?: string
  }) => {
    try {
      const token = localStorage.getItem("authToken")
      const headers: HeadersInit = {
        "Content-Type": "application/json",
      }

      if (token) {
        headers["Authorization"] = `Bearer ${token}`
      }

      // Build query parameters
      const queryParams = new URLSearchParams()
      if (filters?.startDate) queryParams.append("start_date", filters.startDate)
      if (filters?.endDate) queryParams.append("end_date", filters.endDate)
      if (filters?.productId) queryParams.append("product_id", filters.productId)
      if (filters?.cashier) queryParams.append("user", filters.cashier)

      const response = await fetch(`${API_BASE_URL}/api/transactions?${queryParams.toString()}`, {
        headers,
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const result = await response.json()

      if (result.status !== "success") {
        throw new Error(result.message || "Failed to fetch transactions")
      }

      return {
        data: result.data,
        loading: false,
        error: null,
        pagination: result.pagination,
      }
    } catch (error: any) {
      console.error("Error fetching transactions:", error.message)
      return {
        data: [],
        loading: false,
        error: error.message || "Failed to fetch transactions",
      }
    }
  },

  create: async (transaction: {
    items: CartItem[]
    total: number
    user?: string
    paymentMethod: string
    notes?: string
    cashier: string
    customer_name?: string
    customer_email?: string
    customer_phone?: string
  }) => {
    try {
      const token = localStorage.getItem("authToken")
      const headers: HeadersInit = {
        "Content-Type": "application/json",
      }

      if (token) {
        headers["Authorization"] = `Bearer ${token}`
      }

      // Transform the transaction data to match the API's expected format
      const transactionData: CreateTransactionData = {
        items: transaction.items.map((item) => ({
          product_id: item.id,
          quantity: item.quantity,
          price: item.harga_jual,
        })),
        total_price: transaction.total,
        payment_method: transaction.paymentMethod,
        notes: transaction.notes || "",
        user: transaction.user || transaction.cashier, // Use user if provided, otherwise use cashier
        customer_name: transaction.customer_name || undefined,
        customer_email: transaction.customer_email || undefined,
        customer_phone: transaction.customer_phone || undefined,
      }

      const response = await fetch(`${API_BASE_URL}/api/transactions`, {
        method: "POST",
        headers,
        body: JSON.stringify(transactionData),
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const result = await response.json()

      if (result.status !== "success") {
        throw new Error(result.message || "Failed to create transaction")
      }

      return {
        data: result.data,
        loading: false,
        error: null,
      }
    } catch (error: any) {
      console.error("Error creating transaction:", error.message)
      return {
        data: null,
        loading: false,
        error: error.message || "Failed to create transaction",
      }
    }
  },

  getById: async (id: string) => {
    try {
      const token = localStorage.getItem("authToken")
      const headers: HeadersInit = {
        "Content-Type": "application/json",
      }

      if (token) {
        headers["Authorization"] = `Bearer ${token}`
      }

      const response = await fetch(`${API_BASE_URL}/api/transactions/${id}`, {
        headers,
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const result = await response.json()

      if (result.status !== "success") {
        throw new Error(result.message || "Failed to fetch transaction")
      }

      return {
        data: result.data,
        loading: false,
        error: null,
      }
    } catch (error: any) {
      console.error("Error fetching transaction:", error.message)
      return {
        data: null,
        loading: false,
        error: error.message || "Failed to fetch transaction",
      }
    }
  },
}

// Reports API - Updated to use the actual API response
export const reportsApi = {
  getSalesData: async (filters?: {
    startDate?: string
    endDate?: string
    period?: "day" | "month" | "year"
  }) => {
    try {
      const token = localStorage.getItem("authToken")
      const headers: HeadersInit = {
        "Content-Type": "application/json",
      }

      if (token) {
        headers["Authorization"] = `Bearer ${token}`
      }

      // Build query parameters
      const queryParams = new URLSearchParams()
      if (filters?.startDate) queryParams.append("start_date", filters.startDate)
      if (filters?.endDate) queryParams.append("end_date", filters.endDate)
      if (filters?.period) queryParams.append("period", filters.period)

      const response = await fetch(`${API_BASE_URL}/api/reports?${queryParams.toString()}`, {
        headers,
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const result = await response.json()

      if (result.status !== "success") {
        throw new Error(result.message || "Failed to fetch sales report")
      }

      return {
        data: result.data,
        loading: false,
        error: null,
      }
    } catch (error: any) {
      console.error("Error fetching sales report:", error.message)
      return {
        data: {
          average_transaction: 0,
          details: [],
          items_sold: 0,
          total_revenue: 0,
        },
        loading: false,
        error: error.message || "Failed to fetch sales report",
      }
    }
  },
}

// Users API
export const usersApi = {
  getAll: async () => {
    return apiCall<User[]>("/api/users", {}, [{ id: "1", username: "admin", role: "admin" }])
  },
}

export const discountCampaignsApi = {
  getAll: async (page = 1, limit = 10) => {
    try {
      const token = localStorage.getItem("authToken")
      if (!token) throw new Error("Authentication required")
      const response = await fetch(
        `${API_BASE_URL}/api/discount-campaigns?page=${page}&limit=${limit}`,
        { headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` } }
      )
      if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
      const result = await response.json()
      if (result.status !== "success") throw new Error(result.message || "Request failed")
      return { data: result.data as DiscountCampaign[], pagination: result.pagination, error: null }
    } catch (error: any) {
      return { data: [] as DiscountCampaign[], pagination: undefined, error: error.message }
    }
  },

  getById: async (id: string) => {
    try {
      const token = localStorage.getItem("authToken")
      if (!token) throw new Error("Authentication required")
      const response = await fetch(`${API_BASE_URL}/api/discount-campaigns/${id}`, {
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` }
      })
      if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
      const result = await response.json()
      if (result.status !== "success") throw new Error(result.message || "Request failed")
      return { data: result.data as DiscountCampaign, error: null }
    } catch (error: any) {
      return { data: null, error: error.message }
    }
  },

  create: async (data: CreateDiscountCampaignData) => {
    try {
      const token = localStorage.getItem("authToken")
      if (!token) throw new Error("Authentication required")
      const response = await fetch(`${API_BASE_URL}/api/discount-campaigns`, {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
        body: JSON.stringify(data),
      })
      if (!response.ok) {
        const err = await response.json().catch(() => ({}))
        throw new Error(err.message || `HTTP error! status: ${response.status}`)
      }
      const result = await response.json()
      if (result.status !== "success") throw new Error(result.message)
      return { data: result.data as DiscountCampaign, error: null }
    } catch (error: any) {
      return { data: null, error: error.message }
    }
  },

  update: async (id: string, data: UpdateDiscountCampaignData) => {
    try {
      const token = localStorage.getItem("authToken")
      if (!token) throw new Error("Authentication required")
      const response = await fetch(`${API_BASE_URL}/api/discount-campaigns/${id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
        body: JSON.stringify(data),
      })
      if (!response.ok) {
        const err = await response.json().catch(() => ({}))
        throw new Error(err.message || `HTTP error! status: ${response.status}`)
      }
      const result = await response.json()
      if (result.status !== "success") throw new Error(result.message)
      return { data: result.data as DiscountCampaign, error: null }
    } catch (error: any) {
      return { data: null, error: error.message }
    }
  },

  delete: async (id: string) => {
    try {
      const token = localStorage.getItem("authToken")
      if (!token) throw new Error("Authentication required")
      const response = await fetch(`${API_BASE_URL}/api/discount-campaigns/${id}`, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${token}` },
      })
      if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
      return { error: null }
    } catch (error: any) {
      return { error: error.message }
    }
  },

  addProducts: async (campaignId: string, productIds: string[]) => {
    try {
      const token = localStorage.getItem("authToken")
      if (!token) throw new Error("Authentication required")
      const response = await fetch(`${API_BASE_URL}/api/discount-campaigns/${campaignId}/products`, {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
        body: JSON.stringify({ product_ids: productIds }),
      })
      if (!response.ok) {
        const err = await response.json().catch(() => ({}))
        throw new Error(err.message || `HTTP error! status: ${response.status}`)
      }
      return { error: null }
    } catch (error: any) {
      return { error: error.message }
    }
  },

  removeProduct: async (campaignId: string, productId: string) => {
    try {
      const token = localStorage.getItem("authToken")
      if (!token) throw new Error("Authentication required")
      const response = await fetch(
        `${API_BASE_URL}/api/discount-campaigns/${campaignId}/products/${productId}`,
        { method: "DELETE", headers: { Authorization: `Bearer ${token}` } }
      )
      if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
      return { error: null }
    } catch (error: any) {
      return { error: error.message }
    }
  },
}

// Public warranty endpoints - no authentication required
export const warrantyApi = {
  checkByTransactionId: async (transactionId: string) => {
    try {
      const response = await fetch(`${API_BASE_URL}/warranty/${transactionId}`)
      if (!response.ok) {
        const err = await response.json().catch(() => ({}))
        throw new Error(err.message || `HTTP error! status: ${response.status}`)
      }
      const result = await response.json()
      if (result.status !== "success") throw new Error(result.message)
      return { data: result.data as WarrantyResponse, error: null }
    } catch (error: any) {
      return { data: null, error: error.message }
    }
  },

  searchByPhone: async (phone: string) => {
    try {
      const response = await fetch(
        `${API_BASE_URL}/warranty/search?phone=${encodeURIComponent(phone)}`
      )
      if (!response.ok) {
        const err = await response.json().catch(() => ({}))
        throw new Error(err.message || `HTTP error! status: ${response.status}`)
      }
      const result = await response.json()
      if (result.status !== "success") throw new Error(result.message)
      return { data: result.data.transactions as WarrantyResponse[], error: null }
    } catch (error: any) {
      return { data: null, error: error.message }
    }
  },
}

// Tenant API - New service for fetching tenant details
export const tenantApi = {
  getTenantDetails: async () => {
    try {
      const token = localStorage.getItem("authToken")
      if (!token) {
        throw new Error("Authentication token not found. Please log in.")
      }

      const response = await fetch(`${API_BASE_URL}/api/my-tenant`, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        throw new Error(errorData.message || `HTTP error! status: ${response.status}`)
      }

      const result = await response.json()

      if (result.status !== "success") {
        throw new Error(result.message || "Failed to fetch tenant details")
      }

      return { data: result.data, loading: false, error: null }
    } catch (error: any) {
      console.error("Error fetching tenant details:", error.message)
      return {
        data: null,
        loading: false,
        error: error.message || "Failed to fetch tenant details. Please try again.",
      }
    }
  },
}

// Generic API call function with fallback
async function apiCall<T>(
  endpoint: string,
  options: RequestInit = {},
  fallbackData: T,
): Promise<{
  data: T
  loading: boolean
  error: string | null
  pagination?: { total: number; page: number; limit: number }
}> {
  try {
    const token = localStorage.getItem("authToken")

    // Don't make the API call if there's no token
    if (!token) {
      console.warn(`No auth token found, skipping API call to ${endpoint}`)
      return {
        data: fallbackData,
        loading: false,
        error: "Authentication required. Please log in.",
      }
    }

    const headers = {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
      ...options.headers,
    }

    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      headers,
      ...options,
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      const errorMessage = errorData.message || `HTTP error! status: ${response.status}`
      throw new Error(errorMessage)
    }

    const result: ApiResponse<T> = await response.json()

    if (result.status !== "success") {
      throw new Error(result.message || "API call failed")
    }

    return {
      data: result.data,
      loading: false,
      error: null,
      pagination: result.pagination,
    }
  } catch (error: any) {
    console.warn(`API call failed for ${endpoint}, using fallback data:`, error)
    return { data: fallbackData, loading: false, error: error.message || "Failed to fetch data" }
  }
}

// Authentication API - No mock fallback
export const authApi = {
  login: async (username: string, password: string) => {
    try {
      const response = await fetch(`${API_BASE_URL}/auth/login`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ username, password }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || `HTTP error! status: ${response.status}`)
      }

      const result = await response.json()

      if (result.status !== "success") {
        throw new Error(result.message || "Login failed")
      }

      return { data: result.data, loading: false, error: null }
    } catch (error: any) {
      console.error("Login failed:", error.message)
      return {
        data: null,
        loading: false,
        error: error.message || "Login failed. Please check your credentials and try again.",
      }
    }
  },

  updatePassword: async (currentPassword: string, newPassword: string) => {
    try {
      const token = localStorage.getItem("authToken")
      if (!token) {
        throw new Error("Authentication token not found. Please log in.")
      }

      const response = await fetch(`${API_BASE_URL}/api/update-password`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        throw new Error(errorData.message || `HTTP error! status: ${response.status}`)
      }

      const result = await response.json()

      if (result.status !== "success") {
        throw new Error(result.message || "Failed to update password")
      }

      return { data: result.data, loading: false, error: null }
    } catch (error: any) {
      console.error("Password update failed:", error.message)
      return {
        data: null,
        loading: false,
        error: error.message || "Failed to update password. Please try again.",
      }
    }
  },
}

export type { Product, CartItem, Transaction, TransactionItem, User, SalesDetail, SalesReport, Tenant, DiscountCampaign, DiscountCampaignProduct, WarrantyItem, WarrantyResponse }
