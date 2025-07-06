import { getConfig } from '@/lib/config';

const config = getConfig();
const API_BASE_URL = config.apiBaseUrl;

export interface Event {
  id: number;
  name: string;
  description: string;
  venue: string;
  start_time: string;
  end_time: string;
  total_tickets: number;
  price: number;
  available_tickets?: number;
}

export interface Ticket {
  id: number;
  event_id: number;
  seat_no: string;
  status: 'available' | 'reserved' | 'sold';
  created_at: string;
  updated_at: string;
}

export interface BookingRequest {
  user_id: number;
  event_id: number;
  quantity: number;
}

export interface BookingResponse {
  id: number;
  booking_ref: string;
  total_amount: number;
  status: string;
  expires_at: string;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

class ApiService {
  private async request<T>(endpoint: string, options?: RequestInit): Promise<ApiResponse<T>> {
    try {
      console.log(`Making API request to: ${API_BASE_URL}${endpoint}`);
      
      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        headers: {
          'Content-Type': 'application/json',
          ...options?.headers,
        },
        ...options,
      });

      console.log(`API Response status: ${response.status}`);
      
      if (!response.ok) {
        console.error(`API request failed with status: ${response.status}`);
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      console.log(`API Response data:`, data);
      return data;
    } catch (error) {
      console.error('API request failed:', error);
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Network error occurred',
      };
    }
  }

  async getEvents(page = 1, limit = 20): Promise<ApiResponse<Event[]>> {
    return this.request<Event[]>(`/events?page=${page}&limit=${limit}`);
  }

  async getEvent(id: number): Promise<ApiResponse<Event>> {
    return this.request<Event>(`/events/${id}`);
  }

  async getAvailableTickets(eventId: number, limit = 100): Promise<ApiResponse<Ticket[]>> {
    return this.request<Ticket[]>(`/events/${eventId}/tickets?limit=${limit}`);
  }

  async getAllTickets(eventId: number, limit = 100): Promise<ApiResponse<Ticket[]>> {
    return this.request<Ticket[]>(`/events/${eventId}/tickets/all?limit=${limit}`);
  }

  async bookTickets(booking: BookingRequest): Promise<ApiResponse<BookingResponse>> {
    return this.request<BookingResponse>('/bookings', {
      method: 'POST',
      body: JSON.stringify(booking),
    });
  }

  async getBooking(id: number): Promise<ApiResponse<any>> {
    return this.request(`/bookings/${id}`);
  }

  async confirmBooking(id: number): Promise<ApiResponse<any>> {
    return this.request(`/bookings/${id}/confirm`, {
      method: 'POST',
    });
  }

  async cancelBooking(id: number): Promise<ApiResponse<any>> {
    return this.request(`/bookings/${id}/cancel`, {
      method: 'POST',
    });
  }

  async lockSeat(eventId: number, seatNo: string): Promise<ApiResponse<any>> {
    return this.request(`/events/${eventId}/seats/${seatNo}/lock`, {
      method: 'POST',
      headers: {
        'X-Session-ID': `session_${Date.now()}_${Math.random()}`, // Generate session ID
      },
    });
  }

  async unlockSeat(eventId: number, seatNo: string): Promise<ApiResponse<any>> {
    return this.request(`/events/${eventId}/seats/${seatNo}/unlock`, {
      method: 'POST',
    });
  }
}

export const apiService = new ApiService();
