import { memo } from 'react';

function formatTime(timestamp) {
  const date = new Date(timestamp);
  return date.toLocaleTimeString('en-US', {
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

function ErrorListInner({ errors }) {
  return (
    <div className="bg-slate-800 rounded-lg p-6">
      <h2 className="text-lg font-semibold text-white mb-4">Recent Errors</h2>
      <div className="h-48 overflow-y-auto">
        {errors && errors.length > 0 ? (
          <div className="space-y-2">
            {errors.slice().reverse().map((error, index) => (
              <div
                key={error.timestamp + '-' + index}
                className="p-2 bg-red-900/20 border border-red-800/50 rounded text-sm"
              >
                <div className="flex items-start gap-2">
                  <span className="text-red-400 font-mono text-xs whitespace-nowrap">
                    {formatTime(error.timestamp)}
                  </span>
                  <span className="text-red-300 font-mono text-xs break-all">
                    {error.message}
                  </span>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="h-full flex items-center justify-center text-slate-500">
            No errors
          </div>
        )}
      </div>
    </div>
  );
}

export const ErrorList = memo(ErrorListInner);
