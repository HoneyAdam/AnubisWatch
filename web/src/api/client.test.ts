import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { api } from './client'

describe('ApiClient', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.restoreAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('GET', () => {
    it('fetches data successfully', async () => {
      const mockData = { id: '1', name: 'test' }
      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockData),
      })

      const result = await api.get('/test')
      expect(result).toEqual(mockData)
      expect(fetch).toHaveBeenCalledWith(
        '/api/v1/test',
        expect.objectContaining({ method: 'GET' })
      )
    })

    it('throws on 404 error', async () => {
      global.fetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 404,
        json: () => Promise.resolve({ error: 'Not found' }),
      })

      await expect(api.get('/missing')).rejects.toThrow('Not found')
    })

    it('throws on 500 error with fallback message', async () => {
      global.fetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        json: () => Promise.resolve({}),
      })

      await expect(api.get('/error')).rejects.toThrow('HTTP 500')
    })

    it('clears token and redirects to login on 401', async () => {
      Object.defineProperty(window, 'location', {
        value: { href: '' },
        writable: true,
      })

      global.fetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        json: () => Promise.resolve({ error: 'Unauthorized' }),
      })

      api.setToken('test-token')
      await expect(api.get('/protected')).rejects.toThrow('Unauthorized')
      expect(localStorage.getItem('auth_token')).toBeNull()
      expect(window.location.href).toBe('/login')
    })

    it('includes Authorization header when token is set', async () => {
      api.setToken('my-token')
      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve({}),
      })

      await api.get('/data')
      expect(fetch).toHaveBeenCalledWith(
        '/api/v1/data',
        expect.objectContaining({
          headers: expect.objectContaining({
            Authorization: 'Bearer my-token',
          }),
        })
      )
    })

    it('does not include Authorization header when no token', async () => {
      api.clearToken()
      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve({}),
      })

      await api.get('/public')
      expect(fetch).toHaveBeenCalledWith(
        '/api/v1/public',
        expect.objectContaining({
          headers: { 'Content-Type': 'application/json' },
        })
      )
    })
  })

  describe('POST', () => {
    it('sends JSON body', async () => {
      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 201,
        json: () => Promise.resolve({ id: '1' }),
      })

      await api.post('/items', { name: 'new item' })
      expect(fetch).toHaveBeenCalledWith(
        '/api/v1/items',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({ name: 'new item' }),
        })
      )
    })

    it('returns undefined on 204 response', async () => {
      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 204,
        json: () => Promise.resolve({}),
      })

      const result = await api.post('/delete', {})
      expect(result).toBeUndefined()
    })
  })

  describe('PUT', () => {
    it('sends update request', async () => {
      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve({ id: '1', updated: true }),
      })

      const result = await api.put('/items/1', { enabled: false })
      expect(result).toEqual({ id: '1', updated: true })
      expect(fetch).toHaveBeenCalledWith(
        '/api/v1/items/1',
        expect.objectContaining({
          method: 'PUT',
          body: JSON.stringify({ enabled: false }),
        })
      )
    })
  })

  describe('DELETE', () => {
    it('sends delete request', async () => {
      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 204,
        json: () => Promise.resolve({}),
      })

      await api.delete('/items/1')
      expect(fetch).toHaveBeenCalledWith(
        '/api/v1/items/1',
        expect.objectContaining({ method: 'DELETE' })
      )
    })
  })

  describe('token management', () => {
    it('setToken stores in localStorage', () => {
      api.setToken('abc123')
      expect(localStorage.getItem('auth_token')).toBe('abc123')
    })

    it('clearToken removes from localStorage', () => {
      localStorage.setItem('auth_token', 'old-token')
      api.clearToken()
      expect(localStorage.getItem('auth_token')).toBeNull()
    })
  })
})
