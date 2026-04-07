@extends('layouts::docs')

@section('title', 'Stoplight')
@section('description', 'Stoplight Elements API reference for the Core API.')

@section('content')
    <div class="min-h-[calc(100vh-4rem)]">
        <elements-api
            apiDescriptionUrl="{{ route('api.openapi.json') }}"
            router="hash"
            layout="sidebar"
            theme="dark"
        ></elements-api>
    </div>
@endsection

@push('head')
    <link rel="stylesheet" href="https://unpkg.com/@stoplight/elements/styles.min.css">
@endpush

@push('scripts')
    <script src="https://unpkg.com/@stoplight/elements/web-components.min.js"></script>
@endpush
