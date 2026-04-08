import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { Layout } from '../components/Layout'

// Mock useAuth hook
vi.mock('../api/hooks', () => ({
  useAuth: () => ({
    user: { name: 'Test User', email: 'test@anubis.watch' },
    logout: vi.fn(),
  }),
}))

// Mock child components
const MockDashboard = () => <div data-testid="dashboard-content">Dashboard Content</div>

describe('Layout', () => {
  it('renders sidebar and header', () => {
    render(
      <MemoryRouter>
        <Layout />
      </MemoryRouter>
    )

    expect(screen.getByText('Anubis')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Search the archives...')).toBeInTheDocument()
  })

  it('renders outlet content area', () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <Routes>
          <Route path="/" element={<Layout />}>
            <Route index element={<MockDashboard />} />
          </Route>
        </Routes>
      </MemoryRouter>
    )

    expect(screen.getByTestId('dashboard-content')).toBeInTheDocument()
  })

  it('has proper layout structure', () => {
    const { container } = render(
      <MemoryRouter>
        <Layout />
      </MemoryRouter>
    )

    // Check for main layout elements
    const main = container.querySelector('main')
    expect(main).toBeInTheDocument()

    const header = container.querySelector('header')
    expect(header).toBeInTheDocument()

    const aside = container.querySelector('aside')
    expect(aside).toBeInTheDocument()
  })
})
