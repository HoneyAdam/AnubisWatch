import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useSoulStore } from './soulStore'

const mockGet = vi.fn()
const mockPost = vi.fn()
const mockPut = vi.fn()
const mockDelete = vi.fn()

vi.mock('../api/client', () => ({
  api: {
    get: (...args: any[]) => mockGet(...args),
    post: (...args: any[]) => mockPost(...args),
    put: (...args: any[]) => mockPut(...args),
    delete: (...args: any[]) => mockDelete(...args),
  },
}))

describe('useSoulStore', () => {
  beforeEach(() => {
    useSoulStore.setState({
      souls: [],
      pagination: null,
      loading: false,
      error: null,
    })
    mockGet.mockClear()
    mockPost.mockClear()
    mockPut.mockClear()
    mockDelete.mockClear()
  })

  it('fetchSouls loads souls successfully', async () => {
    const souls = [
      { id: '1', name: 'Soul 1', type: 'http', target: 'https://a.com', enabled: true, weight: 1, timeout: 5 },
      { id: '2', name: 'Soul 2', type: 'tcp', target: '10.0.0.1:80', enabled: true, weight: 1, timeout: 5 },
    ]
    mockGet.mockResolvedValue({ data: souls, pagination: { total: 2, has_more: false } })

    await useSoulStore.getState().fetchSouls()

    const state = useSoulStore.getState()
    expect(state.souls).toEqual(souls)
    expect(state.loading).toBe(false)
    expect(state.error).toBeNull()
    expect(state.pagination).toEqual({ total: 2, has_more: false })
  })

  it('fetchSouls sets error on failure', async () => {
    mockGet.mockRejectedValue(new Error('Network error'))

    await useSoulStore.getState().fetchSouls()

    const state = useSoulStore.getState()
    expect(state.loading).toBe(false)
    expect(state.error).toBe('Network error')
    expect(state.souls).toEqual([])
  })

  it('fetchSouls handles non-Error rejection', async () => {
    mockGet.mockRejectedValue('unexpected')

    await useSoulStore.getState().fetchSouls()

    const state = useSoulStore.getState()
    expect(state.error).toBe('Unknown error')
  })

  it('createSoul appends new soul', async () => {
    const newSoul = { id: '3', name: 'Soul 3', type: 'http', target: 'https://c.com', enabled: true, weight: 1, timeout: 5 }
    mockPost.mockResolvedValue(newSoul)

    const result = await useSoulStore.getState().createSoul({
      name: 'Soul 3',
      type: 'http',
      target: 'https://c.com',
      enabled: true,
      weight: 1,
      timeout: 5,
    })

    expect(result).toEqual(newSoul)
    expect(useSoulStore.getState().souls).toContainEqual(newSoul)
    expect(useSoulStore.getState().loading).toBe(false)
  })

  it('createSoul returns null on failure', async () => {
    mockPost.mockRejectedValue(new Error('Create failed'))

    const result = await useSoulStore.getState().createSoul({
      name: 'Soul 3',
      type: 'http',
      target: 'https://c.com',
      enabled: true,
      weight: 1,
      timeout: 5,
    })

    expect(result).toBeNull()
    expect(useSoulStore.getState().error).toBe('Create failed')
  })

  it('updateSoul updates existing soul', async () => {
    useSoulStore.setState({
      souls: [{ id: '1', name: 'Soul 1', type: 'http', target: 'https://a.com', enabled: true, weight: 1, timeout: 5 }],
    })
    const updatedSoul = { id: '1', name: 'Soul 1 Updated', type: 'http', target: 'https://a.com', enabled: true, weight: 1, timeout: 5 }
    mockPut.mockResolvedValue(updatedSoul)

    const result = await useSoulStore.getState().updateSoul('1', { name: 'Soul 1 Updated' })

    expect(result).toEqual(updatedSoul)
    expect(useSoulStore.getState().souls[0].name).toBe('Soul 1 Updated')
  })

  it('updateSoul returns null on failure', async () => {
    mockPut.mockRejectedValue(new Error('Update failed'))

    const result = await useSoulStore.getState().updateSoul('1', { name: 'X' })

    expect(result).toBeNull()
    expect(useSoulStore.getState().error).toBe('Update failed')
  })

  it('deleteSoul removes soul from list', async () => {
    useSoulStore.setState({
      souls: [
        { id: '1', name: 'Soul 1', type: 'http', target: 'https://a.com', enabled: true, weight: 1, timeout: 5 },
        { id: '2', name: 'Soul 2', type: 'tcp', target: '10.0.0.1:80', enabled: true, weight: 1, timeout: 5 },
      ],
    })
    mockDelete.mockResolvedValue(undefined)

    await useSoulStore.getState().deleteSoul('1')

    expect(useSoulStore.getState().souls).toHaveLength(1)
    expect(useSoulStore.getState().souls[0].id).toBe('2')
    expect(useSoulStore.getState().loading).toBe(false)
  })

  it('deleteSoul sets error on failure', async () => {
    mockDelete.mockRejectedValue(new Error('Delete failed'))

    await useSoulStore.getState().deleteSoul('1')

    expect(useSoulStore.getState().error).toBe('Delete failed')
  })
})
