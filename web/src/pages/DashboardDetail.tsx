import { useEffect, useState, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ArrowLeft, Plus, Trash2, RefreshCw, Play } from 'lucide-react'
import { api } from '../api/client'
import type { CustomDashboard, WidgetConfig } from '../api/client'
import { StatWidget } from '../components/widgets/StatWidget'
import { LineChartWidget } from '../components/widgets/LineChartWidget'
import { BarChartWidget } from '../components/widgets/BarChartWidget'
import { GaugeWidget } from '../components/widgets/GaugeWidget'
import { TableWidget } from '../components/widgets/TableWidget'

export function DashboardDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [dashboard, setDashboard] = useState<CustomDashboard | null>(null)
  const [loading, setLoading] = useState(true)
  const [editing, setEditing] = useState(false)
  const [refreshing, setRefreshing] = useState(false)
  const [showAddWidget, setShowAddWidget] = useState(false)

  const fetchDashboard = useCallback(async () => {
    if (!id) return
    setLoading(true)
    try {
      const result = await api.get<CustomDashboard>(`/dashboards/${id}`)
      setDashboard(result)
    } catch {
      setDashboard(null)
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => { fetchDashboard() }, [fetchDashboard])

  // Auto-refresh
  useEffect(() => {
    if (!dashboard || dashboard.refresh_sec <= 0) return
    const interval = setInterval(fetchDashboard, dashboard.refresh_sec * 1000)
    return () => clearInterval(interval)
  }, [dashboard, dashboard?.refresh_sec, fetchDashboard])

  const handleRefresh = async () => {
    setRefreshing(true)
    await fetchDashboard()
    setTimeout(() => setRefreshing(false), 500)
  }

  const handleDeleteWidget = async (widgetId: string) => {
    if (!dashboard || !id) return
    const updated = {
      ...dashboard,
      widgets: dashboard.widgets.filter(w => w.id !== widgetId),
    }
    await api.put(`/dashboards/${id}`, updated)
    await fetchDashboard()
  }

  const handleAddWidget = async (widget: Omit<WidgetConfig, 'id'>) => {
    if (!dashboard || !id) return
    const newWidget: WidgetConfig = {
      ...widget,
      id: `w_${Date.now()}`,
    }
    const updated = {
      ...dashboard,
      widgets: [...dashboard.widgets, newWidget],
    }
    await api.put(`/dashboards/${id}`, updated)
    await fetchDashboard()
    setShowAddWidget(false)
  }

  const renderWidget = (widget: WidgetConfig) => {
    const props = { widget, dashboardId: id! }
    switch (widget.type) {
      case 'stat': return <StatWidget {...props} />
      case 'line_chart': return <LineChartWidget {...props} />
      case 'bar_chart': return <BarChartWidget {...props} />
      case 'gauge': return <GaugeWidget {...props} />
      case 'table': return <TableWidget {...props} />
      default: return <div className="text-gray-500">Unknown widget type</div>
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="w-8 h-8 border-2 border-amber-500/30 border-t-amber-500 rounded-full animate-spin" />
      </div>
    )
  }

  if (!dashboard) {
    return (
      <div className="text-center py-16">
        <h3 className="text-2xl font-bold text-white mb-2">Dashboard Not Found</h3>
        <button onClick={() => navigate('/dashboards')} className="text-amber-400 hover:text-amber-300">
          Back to Dashboards
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate('/dashboards')} className="text-gray-400 hover:text-amber-400 transition-colors">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div>
            <h1 className="text-3xl font-cinzel font-bold gradient-gold-shine tracking-wider">{dashboard.name}</h1>
            {dashboard.description && (
              <p className="text-gray-400 font-cormorant italic mt-1">{dashboard.description}</p>
            )}
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={handleRefresh}
            className={`p-2 bg-[#D4AF37]/10 hover:bg-[#D4AF37]/20 text-[#D4AF37] rounded-lg border border-[#D4AF37]/30 transition-all ${refreshing ? 'animate-spin' : ''}`}
          >
            <RefreshCw className="w-4 h-4" />
          </button>
          <button
            onClick={() => setEditing(!editing)}
            className={`px-3 py-2 rounded-lg border text-sm font-medium transition-all ${
              editing
                ? 'bg-amber-500/20 border-amber-500/40 text-amber-400'
                : 'bg-gray-800/50 border-gray-700/50 text-gray-400 hover:text-white'
            }`}
          >
            {editing ? 'Done Editing' : 'Edit'}
          </button>
          {editing && (
            <button
              onClick={() => setShowAddWidget(!showAddWidget)}
              className="flex items-center gap-1 px-3 py-2 bg-amber-500/20 border border-amber-500/40 text-amber-400 rounded-lg text-sm font-medium hover:bg-amber-500/30 transition-all"
            >
              <Plus className="w-4 h-4" />
              Add Widget
            </button>
          )}
        </div>
      </div>

      {/* Add Widget Form */}
      {showAddWidget && editing && (
        <AddWidgetForm onAdd={handleAddWidget} onCancel={() => setShowAddWidget(false)} />
      )}

      {/* Widget Grid */}
      {dashboard.widgets.length === 0 ? (
        <div className="text-center py-16 bg-gray-900/50 rounded-2xl border border-gray-700/50">
          <p className="text-gray-400 mb-4">No widgets yet</p>
          {editing && (
            <button
              onClick={() => setShowAddWidget(true)}
              className="px-4 py-2 bg-amber-500/20 border border-amber-500/40 text-amber-400 rounded-lg text-sm font-medium"
            >
              Add Your First Widget
            </button>
          )}
        </div>
      ) : (
        <div className="relative" style={{ display: 'grid', gridTemplateColumns: 'repeat(12, 1fr)', gap: '1rem', gridAutoRows: 'minmax(80px, auto)' }}>
          {dashboard.widgets.map((widget) => (
            <div
              key={widget.id}
              className="bg-gradient-to-br from-gray-900 to-gray-800/50 border border-gray-700/50 rounded-2xl p-4 overflow-hidden"
              style={{
                gridColumn: `span ${widget.grid.width}`,
                gridRow: `${widget.grid.y + 1} / span ${widget.grid.height}`,
              }}
            >
              {editing && (
                <div className="flex items-center justify-between mb-2">
                  <span className="text-xs text-gray-500">{widget.type}</span>
                  <button
                    onClick={() => handleDeleteWidget(widget.id)}
                    className="text-gray-500 hover:text-red-400 transition-colors"
                  >
                    <Trash2 className="w-3 h-3" />
                  </button>
                </div>
              )}
              {!editing && (
                <h3 className="text-sm font-medium text-gray-400 mb-2">{widget.title}</h3>
              )}
              <div className={editing ? '' : 'h-full'} style={{ height: editing ? '200px' : `${widget.grid.height * 80 - 40}px` }}>
                {renderWidget(widget)}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

// Add Widget Form component
function AddWidgetForm({ onAdd, onCancel }: { onAdd: (w: Omit<WidgetConfig, 'id'>) => void; onCancel: () => void }) {
  const [title, setTitle] = useState('')
  const [type, setType] = useState<WidgetConfig['type']>('stat')
  const [source, setSource] = useState('souls')
  const [metric, setMetric] = useState('count')
  const [timeRange, setTimeRange] = useState('24h')
  const [width, setWidth] = useState(4)
  const [height, setHeight] = useState(2)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onAdd({
      title: title || 'Untitled',
      type,
      grid: { x: 0, y: 0, width, height },
      query: { source, metric, time_range: timeRange },
    })
  }

  return (
    <form onSubmit={handleSubmit} className="bg-gray-900/80 border border-amber-500/30 rounded-2xl p-6 space-y-4">
      <h3 className="text-lg font-semibold text-white">Add Widget</h3>
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div>
          <label className="text-xs text-gray-400 block mb-1">Title</label>
          <input value={title} onChange={e => setTitle(e.target.value)} className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white text-sm" placeholder="Widget title" />
        </div>
        <div>
          <label className="text-xs text-gray-400 block mb-1">Type</label>
          <select value={type} onChange={e => setType(e.target.value as WidgetConfig['type'])} className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white text-sm">
            <option value="stat">Stat</option>
            <option value="line_chart">Line Chart</option>
            <option value="bar_chart">Bar Chart</option>
            <option value="gauge">Gauge</option>
            <option value="table">Table</option>
          </select>
        </div>
        <div>
          <label className="text-xs text-gray-400 block mb-1">Source</label>
          <select value={source} onChange={e => setSource(e.target.value)} className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white text-sm">
            <option value="souls">Souls</option>
            <option value="judgments">Judgments</option>
            <option value="stats">Stats</option>
            <option value="alerts">Alerts</option>
          </select>
        </div>
        <div>
          <label className="text-xs text-gray-400 block mb-1">Metric</label>
          <input value={metric} onChange={e => setMetric(e.target.value)} className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white text-sm" placeholder="metric" />
        </div>
        <div>
          <label className="text-xs text-gray-400 block mb-1">Width (1-12)</label>
          <input type="number" min={1} max={12} value={width} onChange={e => setWidth(parseInt(e.target.value))} className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white text-sm" />
        </div>
        <div>
          <label className="text-xs text-gray-400 block mb-1">Height (rows)</label>
          <input type="number" min={1} max={8} value={height} onChange={e => setHeight(parseInt(e.target.value))} className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white text-sm" />
        </div>
      </div>
      <div className="flex gap-2">
        <button type="submit" className="px-4 py-2 bg-amber-500/20 border border-amber-500/40 text-amber-400 rounded-lg text-sm font-medium hover:bg-amber-500/30 transition-all">
          Add Widget
        </button>
        <button type="button" onClick={onCancel} className="px-4 py-2 bg-gray-800 border border-gray-700 text-gray-400 rounded-lg text-sm hover:text-white transition-all">
          Cancel
        </button>
      </div>
    </form>
  )
}
