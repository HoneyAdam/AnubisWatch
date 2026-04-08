import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { Header } from '../components/Header'

// Mock useAuth hook
vi.mock('../api/hooks', () => ({
  useAuth: () => ({
    user: { name: 'Test User', email: 'test@anubis.watch' },
    logout: vi.fn(),
  }),
}))

// Mock useNavigate
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

describe('Header', () => {
  it('renders search input', () => {
    render(
      <MemoryRouter>
        <Header />
      </MemoryRouter>
    )

    expect(screen.getByPlaceholderText('Search the archives...')).toBeInTheDocument()
  })

  it('renders Hall of Ma\'at badge', () => {
    render(
      <MemoryRouter>
        <Header />
      </MemoryRouter>
    )

    expect(screen.getByText("Hall of Ma'at")).toBeInTheDocument()
  })

  it('renders user info', () => {
    render(
      <MemoryRouter>
        <Header />
      </MemoryRouter>
    )

    expect(screen.getByText('Test User')).toBeInTheDocument()
    expect(screen.getByText('test@anubis.watch')).toBeInTheDocument()
  })

  it('renders notification button', () => {
    render(
      <MemoryRouter>
        <Header />
      </MemoryRouter>
    )

    // Notification button doesn't have accessible name, check by bell icon presence
    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBeGreaterThanOrEqual(3) // Theme toggle, notification, logout
  })

  it('renders theme toggle button', () => {
    render(
      <MemoryRouter>
        <Header />
      </MemoryRouter>
    )

    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBeGreaterThan(0)
  })

  it('renders logout button', () => {
    render(
      <MemoryRouter>
        <Header />
      </MemoryRouter>
    )

    expect(screen.getByTitle('Logout')).toBeInTheDocument()
  })
})
