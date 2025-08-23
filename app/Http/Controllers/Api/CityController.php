<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Models\City;
use Illuminate\Http\Request;
use Illuminate\Http\Response;
use Illuminate\Support\Facades\Validator;


class CityController extends Controller
{
    public function index(Request $request)
    {
        $query = City::query();

        if ($request->has('region_id')) {
            $query->where('region_id', $request->region_id);
        }

        return response()->json($query->orderBy('name')->get());
    }
}