import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/vue'
import JobStatus from './JobStatus.vue'

describe('JobStatus', () => {
  it('displays the correct status', () => {
    const status = 'completed'
    render(JobStatus, {
      props: {
        status,
      },
    })

    expect(screen.getByText(status)).toBeTruthy()
  })

  it('applies the correct status class', () => {
    const status = 'running'
    const { container } = render(JobStatus, {
      props: {
        status,
      },
    })

    expect(container.querySelector('.status-running')).toBeTruthy()
  })

  it('emits status-click event when clicked', async () => {
    const status = 'pending'
    const { emitted } = render(JobStatus, {
      props: {
        status,
      },
    })

    const statusElement = screen.getByText(status)
    await statusElement.click()

    expect(emitted()['status-click']).toBeTruthy()
    expect(emitted()['status-click'][0]).toEqual([status])
  })
})
