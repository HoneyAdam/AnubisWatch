import { useEffect, useState } from 'react'
import { Activity, AlertTriangle, CheckCircle2, Clock, Zap, X } from 'lucide-react'
import { formatDistanceToNow } from '../utils/date'

interface Event {
  id: string
  type: 'success' | 'warning' | 'error' | 'info'
  message: string
  soulName?: string
  timestamp: string
}

interface EventsFeedProps {
  maxEvents?: number
}

export function EventsFeed({ maxEvents = 10 }: EventsFeedProps) {
  const [events, setEvents] = useState<Event[]>([])
  const [dismissed, setDismissed] = useState<Set<string>>(new Set())

  // Simulate real-time events (in production, this would come from WebSocket)
  useEffect(() => {
    // Initial events
    setEvents([
      {
        id: '1',
        type: 'success',
        message: 'Health check passed',
        soulName: 'API Server',
        timestamp: new Date(Date.now() - 1000 * 60 * 2).toISOString()
      },
      {
        id: '2',
        type: 'info',
        message: 'Configuration updated',
        timestamp: new Date(Date.now() - 1000 * 60 * 15).toISOString()
      },
      {
        id: '3',
        type: 'warning',
        message: 'High latency detected',
        soulName: 'Database',
        timestamp: new Date(Date.now() - 1000 * 60 * 30).toISOString()
      }
    ])
  }, [])

  const dismissEvent = (id: string) => {
    setDismissed(prev => new Set(prev).add(id))
  }

  const filteredEvents = events.filter(e => !dismissed.has(e.id)).slice(0, maxEvents)

  const getIcon = (type: Event['type']) => {
    switch (type) {
      case 'success':
        return <CheckCircle2 className="w-4 h-4 text-emerald-400" />
      case 'warning':
        return <AlertTriangle className="w-4 h-4 text-amber-400" />
      case 'error':
        return <Zap className="w-4 h-4 text-rose-400" />
      case 'info':
        return <Activity className="w-4 h-4 text-blue-400" />
    }
  }

  const getBorderColor = (type: Event['type']) => {
    switch (type) {
      case 'success':
        return 'border-emerald-500/20 hover:border-emerald-500/40'
      case 'warning':
        return 'border-amber-500/20 hover:border-amber-500/40'
      case 'error':
        return 'border-rose-500/20 hover:border-rose-500/40'
      case 'info':
        return 'border-blue-500/20 hover:border-blue-500/40'
    }
  }

  if (filteredEvents.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500">
        <Clock className="w-8 h-8 mx-auto mb-2 opacity-50" />
        <p className="text-sm">No recent events</p>
      </div>
    )
  }

  return (
    <div className="space-y-2">
      {filteredEvents.map((event) => (
        <div
          key={event.id}
          className={`group flex items-start gap-3 p-3 rounded-lg bg-gray-800/30 border ${getBorderColor(event.type)}
                      transition-all duration-300 animate-slide-in`}
        >
          <div className="mt-0.5">{getIcon(event.type)}</div>
          <div className="flex-1 min-w-0">
            <p className="text-sm text-gray-200">
              {event.message}
              {event.soulName && (
                <span className="text-gray-500"> · {event.soulName}</span>
              )}
            </p>
            <p className="text-xs text-gray-500 mt-0.5">
              {formatDistanceToNow(event.timestamp)}
            </p>
          </div>
          <button
            onClick={() => dismissEvent(event.id)}
            className="opacity-0 group-hover:opacity-100 p-1 text-gray-500 hover:text-gray-300
                       transition-opacity"
            aria-label="Dismiss event"
          >
            <X className="w-3 h-3" />
          </button>
        </div>
      ))}
    </div>
  )
}
