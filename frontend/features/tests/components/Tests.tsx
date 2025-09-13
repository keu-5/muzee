'use client';

import { Providers } from "@/app/providers";
import { useGetTests, usePostTests } from "@/src/api/__generated__/tests/tests";

const _Tests = () => {
    const { data: tests, isLoading, error, refetch } = useGetTests();
    const createTest = usePostTests();

    const handleCreateTest = () => {
        createTest.mutate(undefined, {
        onSuccess: () => {
            refetch();
        }
        });
    };

    return (
        <div className="w-full max-w-2xl p-6 border border-gray-200 dark:border-gray-700 rounded-lg">
            <h2 className="text-xl font-semibold mb-4">API Test Results</h2>

            <div className="space-y-4">
            <div>
                <button
                onClick={handleCreateTest}
                disabled={createTest.isPending}
                className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
                >
                {createTest.isPending ? 'Creating Test...' : 'Create New Test'}
                </button>
                {createTest.error && (
                <p className="text-red-500 mt-2">Error: {createTest.error.message}</p>
                )}
                {createTest.data && (
                <p className="text-green-500 mt-2">Created test with ID: {createTest.data.id}</p>
                )}
            </div>

            <div>
                <h3 className="font-medium mb-2">All Tests:</h3>
                {isLoading && <p>Loading tests...</p>}
                {error && <p className="text-red-500">Error loading tests: {error.message}</p>}
                {tests && (
                <div className="bg-gray-100 dark:bg-gray-800 p-3 rounded">
                    <pre className="text-sm">{JSON.stringify(tests, null, 2)}</pre>
                </div>
                )}
            </div>
            </div>
        </div>
    );
};

export const Tests = () => {
    return (
        <Providers>
            <_Tests />
        </Providers>
    );
};