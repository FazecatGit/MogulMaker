import { render, screen } from '@testing-library/react';
import PageHeader from '../PageHeader';

describe('PageHeader', () => {
  it('renders title correctly', () => {
    render(<PageHeader title="Test Title" />);
    expect(screen.getByText('Test Title')).toBeInTheDocument();
  });

  it('renders description when provided', () => {
    render(<PageHeader title="Test Title" description="Test Description" />);
    expect(screen.getByText('Test Description')).toBeInTheDocument();
  });

  it('does not render description when not provided', () => {
    const { container } = render(<PageHeader title="Test Title" />);
    expect(container.querySelector('p')).not.toBeInTheDocument();
  });

  it('applies correct CSS class', () => {
    const { container } = render(<PageHeader title="Test Title" />);
    expect(container.querySelector('.page-header')).toBeInTheDocument();
  });
});
